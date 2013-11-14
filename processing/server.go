//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//Package processing provides functionality for running a submission and its snapshots
//through the Impendulo tool suite.
package processing

import (
	"container/list"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (

	//Request is used to carry requests to process submissions and files.
	Request struct {
		Id    bson.ObjectId
		Start bool
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
	Empty struct{}

	//ProcHelper is used to help handle a submission's files.
	ProcHelper struct {
		subId     bson.ObjectId
		serveChan chan bson.ObjectId
		doneChan  chan Empty
		started   bool
		done      bool
	}
)

const (
	LOG_SERVER = "processing/server.go"
)

var (
	processedChan chan Empty
	statusChan    chan Status
	active        bool
)

func init() {
	processedChan = make(chan Empty)
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

//monitorStatus keeps track of Impendulo's current processing status.
func monitorStatus() {
	//This is Impendulo's actual processing status.
	var status *Status = new(Status)
	for val := range statusChan {
		switch val {
		case Status{}:
			//A zeroed Status indicates a request for the current Status.
			statusChan <- *status
		default:
			status.add(val)
		}
	}
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
func None() Empty {
	return Empty{}
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func Serve(maxProcs int) {
	if active {
		return
	}
	active = true
	defer func() {
		active = false
	}()
	fileCh := make(chan *Request)
	subCh := make(chan *Request)
	helpers := make(map[bson.ObjectId]*ProcHelper)
	fileQueues := make(map[bson.ObjectId]*list.List)
	subQueue := list.New()
	busy := 0
	//Begin monitoring processing status
	go monitorStatus()
	go func() {
		err := startMQ(AMQP_URI, subCh, fileCh, statusChan)
		if err != nil {
			util.Log(err)
		}
	}()
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
		case request := <-fileCh:
			f, err := db.File(bson.M{db.ID: request.Id}, bson.M{db.SUBID: 1})
			if err != nil {
			} else if helper, ok := helpers[f.SubId]; ok {
				if helper.started {
					//Send file to goroutine if
					//submission processing has started.
					helper.serveChan <- request.Id
				} else {
					//Add file to queue if not.
					fileQueues[f.SubId].PushBack(request.Id)
				}
				ChangeStatus(Status{1, 0})
			} else {
				err = fmt.Errorf("No submission %q found for file %q.",
					f.SubId, request.Id)
			}
			request.Response <- err
		case request := <-subCh:
			var err error
			helper, ok := helpers[request.Id]
			if request.Start && !ok {
				//Add submission to queue.
				subQueue.PushBack(request.Id)
				helpers[request.Id] = NewProcHelper(request.Id)
				fileQueues[request.Id] = list.New()
				ChangeStatus(Status{0, 1})
			} else if !request.Start && ok {
				//Submission will receive no more files.
				helper.SetDone()
				if helper.Started() {
					delete(helpers, request.Id)
				}
			} else if request.Start && ok {
				err = fmt.Errorf("Can't start submission %q, already being processed.", request.Id)
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

//NewProcHelper creates a new ProcHelper for the specified
//Submission.
func NewProcHelper(subId bson.ObjectId) *ProcHelper {
	return &ProcHelper{
		subId:     subId,
		serveChan: make(chan bson.ObjectId),
		doneChan:  make(chan Empty),
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
				<-stopChan
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
