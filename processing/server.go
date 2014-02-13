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
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"runtime"
)

type (

	//Request is used to carry requests to process submissions and files.
	Request struct {
		SubId, FileId bson.ObjectId
		Type          RequestType
	}
	RequestType uint8

	//empty
	E struct{}

	//procHelper is used to help handle a submission's files.
	procHelper struct {
		subId     bson.ObjectId
		serveChan chan bson.ObjectId
		doneChan  chan E
		started   bool
		done      bool
	}

	//Server is our processing server which receives and processes submissions and files.
	Server struct {
		maxProcs      uint
		requestChan   chan *Request
		processedChan chan E
		//submitter listens for messages on AMQP which indicate that a submission has started.
		redoer, starter, submitter *MessageHandler
	}
)

const (
	LOG_SERVER                   = "processing/server.go"
	SUBMISSION_START RequestType = iota
	SUBMISSION_STOP
	FILE_ADD
)

var (
	defaultServer *Server
	MAX_PROCS     = max(runtime.NumCPU()-1, 1)
)

//max is a convenience function to find the largest of two integers.
func max(a, b int) uint {
	if a < 0 {
		a = 0
	}
	if b > a {
		a = b
	}
	return uint(a)
}

//Serve launches the default Server. It listens on the configured AMQP URI and
//spawns at most maxProcs goroutines in order to process submissions.
func Serve(maxProcs uint) (err error) {
	if defaultServer, err = NewServer(maxProcs); err == nil {
		defaultServer.Serve()
	}
	return
}

//Shutdown signals to the default Server that it can shutdown
//and waits for it to complete all processing. It then shuts down all
//active producers as well as the status monitor.
func Shutdown() (err error) {
	err = defaultServer.Shutdown()
	if err != nil {
		return
	}
	err = StopProducers()
	if err != nil {
		return
	}
	err = ShutdownMonitor()
	return
}

//NewServer constructs a new Server instance which will listen on the coinfigured
//AMQP URI.
func NewServer(maxProcs uint) (ret *Server, err error) {
	ret = &Server{
		maxProcs:      maxProcs,
		requestChan:   make(chan *Request),
		processedChan: make(chan E),
	}
	ret.submitter, ret.starter, err = NewSubmitter(ret.requestChan)
	if err != nil {
		return
	}
	ret.redoer, err = NewRedoer(ret.requestChan)
	return
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func (this *Server) Serve() {
	go handleFunc(this.starter)
	go handleFunc(this.submitter)
	go handleFunc(this.redoer)
	helpers := make(map[bson.ObjectId]*procHelper)
	fileQueues := make(map[bson.ObjectId]*list.List)
	subQueue := list.New()
	var busy uint = 0
	//Begin monitoring processing status
	for {
		if busy < this.maxProcs && subQueue.Len() > 0 {
			//If there is an available spot,
			//start processing the next submission.
			subId := subQueue.Remove(subQueue.Front()).(bson.ObjectId)
			helper := helpers[subId]
			helper.started = true
			if helper.done {
				delete(helpers, subId)
			}
			go helper.handle(this.processedChan, fileQueues[subId])
			delete(fileQueues, subId)
			busy++
		} else if busy < 0 {
			//This will only occur when Shutdown() has been called and
			//all submissions have been completed and processed.
			break
		}
		select {
		case request := <-this.requestChan:
			helper, ok := helpers[request.SubId]
			switch request.Type {
			case SUBMISSION_STOP:
				if !ok {
					util.Log(fmt.Errorf("No submission %q found to end.", request.SubId))
				} else {
					//If the submission has finished, set the submission's procHelper to done
					//and if it has already started, remove it from the queue.
					helper.setDone()
					if helper.started {
						delete(helpers, request.SubId)
					}
				}
			case SUBMISSION_START:
				if ok {
					util.Log(fmt.Errorf("Submission %s already started.", request.SubId))
				} else {
					//This is a new submission so we initialise it.
					subQueue.PushBack(request.SubId)
					helpers[request.SubId] = newHelper(request.SubId)
					fileQueues[request.SubId] = list.New()
					ChangeStatus(Status{Files: 0, Submissions: 1})
				}
			case FILE_ADD:
				if !ok {
					util.Log(fmt.Errorf("No submission %s found for file %.", request.SubId, request.FileId))
				} else {
					if helper.started {
						//Send file to goroutine if
						//submission processing has started.
						helper.serveChan <- request.FileId
					} else {
						//Add file to queue if not.
						fileQueues[request.SubId].PushBack(request.FileId)
					}
					ChangeStatus(Status{Files: 1, Submissions: 0})
				}
			default:
				util.Log(fmt.Errorf("Unsupported request type %d.", request.Type))
			}
		case <-this.processedChan:
			//A submission has been processed so one less goroutine to worry about.
			busy--
		}
	}
}

//Shutdown stops Serve from running once all submissions have been processed.
func (this *Server) Shutdown() (err error) {
	this.processedChan <- E{}
	err = this.submitter.Shutdown()
	if err != nil {
		return
	}
	err = this.starter.Shutdown()
	if err != nil {
		return
	}
	err = this.redoer.Shutdown()
	return
}

//newHelper creates a new procHelper for the specified
//Submission.
func newHelper(subId bson.ObjectId) *procHelper {
	return &procHelper{
		subId:     subId,
		serveChan: make(chan bson.ObjectId),
		doneChan:  make(chan E),
		started:   false,
		done:      false,
	}
}

//SetDone indicates that this submission will receive no more files.
func (this *procHelper) setDone() {
	if this.started {
		//If it has started send on its channel.
		this.doneChan <- E{}
	} else {
		//Otherwise simply set done to true.
		this.done = true
	}
}

//handle helps manage the files a submission receives.
//It spawns a new Processor which runs in a seperate goroutine
//and receives files to process from this procHelper.
//fileQueue is the queue of files the submission has received
//prior to the start of processing.
func (this *procHelper) handle(onDone chan E, fileQueue *list.List) {
	procChan := make(chan bson.ObjectId)
	stopChan := make(chan E)
	proc, err := NewProcessor(this.subId)
	if err != nil {
		util.Log(err, LOG_SERVER)
		ChangeStatus(Status{-fileQueue.Len(), -1})
		onDone <- E{}
		return
	}
	go proc.Process(procChan, stopChan)
	busy := false
	for {
		if !busy {
			if fileQueue.Len() > 0 {
				//Not busy and there are files so send one to be processed.
				fId := fileQueue.Remove(fileQueue.Front()).(bson.ObjectId)
				procChan <- fId
				busy = true
			} else if this.done {
				//Not busy and we are done so we should finish up here.
				stopChan <- E{}
				<-stopChan
				ChangeStatus(Status{0, -1})
				onDone <- E{}
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
			ChangeStatus(Status{-1, 0})
		case <-this.doneChan:
			//Submission will receive no more files.
			this.done = true
		}
	}
}
