//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

//Package processing provides functionality for running a submission and its snapshots
//through the Impendulo tool suite.
package processing

import (
	"container/list"
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

const (
	LOG_SERVER = "processing/server.go"
)

var (
	fileChan, startSubmissionChan, endSubmissionChan chan *Request
	processedChan                                    chan interface{}
	statusChan                                       chan Status
)

type (
	//Request is used to carry requests to process submissions and files.
	Request struct {
		Id, ParentId bson.ObjectId
		//Used to return errors
		Response chan error
	}

	//Status is used to indicate a change in the files or
	//submissions being processed. It is also used to retrieve the current
	//number of files and submissions being processed.
	Status struct {
		Files       int
		Submissions int
	}

	//empty
	empty struct{}

	//ProcHelper is used to help handle a submission's files.
	ProcHelper struct {
		subId     bson.ObjectId
		serveChan chan bson.ObjectId
		doneChan  chan interface{}
		started   bool
		done      bool
	}
)

func init() {
	fileChan = make(chan *Request)
	startSubmissionChan = make(chan *Request)
	endSubmissionChan = make(chan *Request)
	processedChan = make(chan interface{})
	statusChan = make(chan Status)
}

//add adds the value of toAdd to this Status.
func (this *Status) add(toAdd Status) {
	this.Files += toAdd.Files
	this.Submissions += toAdd.Submissions
}

//ChangeStatus is used to update Impendulo's current
//processing status.
func ChangeStatus(change Status) {
	statusChan <- change
}

//GetStatus retrieves Impendulo's current processing status.
func GetStatus() (ret Status) {
	statusChan <- Status{0, 0}
	ret = <-statusChan
	return
}

//monitorStatus keeps track of Impendulo's current processing status.
func monitorStatus() {
	//This is Impendulo's actual processing status.
	var status *Status = new(Status)
	for {
		val := <-statusChan
		switch val {
		case Status{0, 0}:
			//A zeroed Status indicates a request for the current Status.
			statusChan <- *status
		default:
			status.add(val)
		}
	}
}

//AddFile sends a file id to be processed.
func AddFile(file *project.File) error {
	//We only need to process source files  and archives.
	if !file.CanProcess() {
		return nil
	}
	errChan := make(chan error)
	//We send the file's db id as well as the id of the submission to which it belongs.
	fileChan <- &Request{
		Id:       file.Id,
		ParentId: file.SubId,
		Response: errChan,
	}
	//Return any errors which occured while adding the file.
	return <-errChan
}

//StartSubmission signals that this submission will now receive files.
func StartSubmission(subId bson.ObjectId) error {
	errChan := make(chan error)
	//We send the submission's db id as a reference.
	startSubmissionChan <- &Request{
		Id:       subId,
		Response: errChan,
	}
	return <-errChan
}

//EndSubmission signals that this submission has stopped receiving files.
func EndSubmission(subId bson.ObjectId) error {
	errChan := make(chan error)
	endSubmissionChan <- &Request{
		Id:       subId,
		Response: errChan,
	}
	return <-errChan
}

//submissionProcessed
func submissionProcessed() {
	ChangeStatus(Status{0, -1})
	processedChan <- None()
}

//fileProcessed
func fileProcessed() {
	ChangeStatus(Status{-1, 0})
}

//Shutdown stops Serve from running once all submissions have been processed.
func Shutdown() {
	processedChan <- None()
}

//None provides an empty struct
func None() interface{} {
	return empty{}
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func Serve(maxProcs int) {
	//Begin monitoring processing status
	go monitorStatus()
	helpers := make(map[bson.ObjectId]*ProcHelper)
	fileQueues := make(map[bson.ObjectId]*list.List)
	subQueue := list.New()
	busy := 0
	for {
		if busy < maxProcs && subQueue.Len() > 0 {
			//If there is an available spot,
			//start processing the next submission.
			subId := subQueue.Remove(subQueue.Front()).(bson.ObjectId)
			helper := helpers[subId]
			helper.SetStarted()
			if helper.Done() {
				delete(helpers, subId)
			}
			go helper.Handle(fileQueues[subId])
			delete(fileQueues, subId)
			busy++
		} else if busy < 0 {
			//This will only occur when Shutdown() has been called and
			//all submissions have been completed and processed.
			break
		}
		select {
		case request := <-fileChan:
			var err error
			if helper, ok := helpers[request.ParentId]; ok {
				if helper.started {
					//Send file to goroutine if
					//submission processing has started.
					helper.serveChan <- request.Id
				} else {
					//Add file to queue if not.
					fileQueues[request.ParentId].PushBack(request.Id)
				}
				ChangeStatus(Status{1, 0})
			} else {
				err = fmt.Errorf("No submission %q found for file %q.",
					request.ParentId, request.Id)
			}
			request.Response <- err
		case request := <-startSubmissionChan:
			var err error
			if _, ok := helpers[request.Id]; !ok {
				//Add submission to queue.
				subQueue.PushBack(request.Id)
				helpers[request.Id] = NewProcHelper(request.Id)
				fileQueues[request.Id] = list.New()
				ChangeStatus(Status{0, 1})
			} else {
				err = fmt.Errorf("Can't start submission %q, already being processed.", request.Id)
			}
			request.Response <- err
		case request := <-endSubmissionChan:
			var err error
			if helper, ok := helpers[request.Id]; ok {
				//Submission will receive no more files.
				helper.SetDone()
				if helper.Started() {
					delete(helpers, request.Id)
				}
			} else {
				err = fmt.Errorf("No submission %q found to end.", request.Id)
			}
			request.Response <- err
		case <-processedChan:
			//A submission has been processed so one less goroutine to worry about.
			busy--
		}
	}
}

//NewProcHelper
func NewProcHelper(subId bson.ObjectId) *ProcHelper {
	return &ProcHelper{
		subId:     subId,
		serveChan: make(chan bson.ObjectId),
		doneChan:  make(chan interface{}),
		started:   false,
		done:      false,
	}
}

//SetDone indicates that this submission will receive no more files.
func (this *ProcHelper) SetDone() {
	if this.Started() {
		//If it has started send on its channel.
		this.doneChan <- None()
	} else {
		//Otherwise simply set done to true.
		this.done = true
	}
}

//Done
func (this *ProcHelper) Done() bool {
	return this.done
}

//SetStarted
func (this *ProcHelper) SetStarted() {
	this.started = true
}

//Started
func (this *ProcHelper) Started() bool {
	return this.started
}

//Handle helps manage the files a submission receives.
//It spawns a new Processor which runs in a seperate goroutine
//and receives files to process from this ProcHelper.
//fileQueue is the queue of files the submission has received
//prior to the start of processing.
func (this *ProcHelper) Handle(fileQueue *list.List) {
	procChan := make(chan bson.ObjectId)
	stopChan := make(chan interface{})
	proc, err := NewProcessor(this.subId)
	if err != nil {
		util.Log(err, LOG_SERVER)
		submissionProcessed()
		return
	}
	go proc.Process(procChan, stopChan)
	busy := false
	for {
		if !busy {
			if fileQueue.Len() > 0 {
				//Not busy and there are files so send one to be processed.
				fId := fileQueue.Remove(
					fileQueue.Front()).(bson.ObjectId)
				procChan <- fId
				busy = true
			} else if this.done {
				//Not busy and we are done so we should finish up here.
				stopChan <- None()
				submissionProcessed()
				return
			}
		}
		select {
		case fId := <-this.serveChan:
			//Add new files to the queue.
			fileQueue.PushBack(fId)
		case <-procChan:
			//Processor has finished with its current file.
			busy = false
			fileProcessed()
		case <-this.doneChan:
			//Submission will receive no more files.
			this.done = true
		}
	}
}
