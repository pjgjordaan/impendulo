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
		Stop          bool
	}

	//empty
	E struct{}

	//ProcHelper is used to help handle a submission's files.
	ProcHelper struct {
		subId     bson.ObjectId
		serveChan chan bson.ObjectId
		doneChan  chan E
		started   bool
		done      bool
	}

	//Server is our processing server which receives and processes submissions and files.
	Server struct {
		uri           string
		maxProcs      int
		reqChan       chan *Request
		processedChan chan E
		//ender listens for messages on AMQP which indicate that a submission has ended.
		//fileConsumer retrieves new files for processing from AMQP.
		ender, fileConsumer *MessageHandler
	}
)

var (
	defaultServer *Server
	MAX_PROCS     = max(runtime.NumCPU()-1, 1)
)

const (
	LOG_SERVER = "processing/server.go"
)

//max is a convenience function to find the largest of two integers.
func max(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

//Serve launches the default Server. It listens on the provided AMQP URI and
//spawns at most maxProcs goroutines in order to process submissions.
func Serve(amqpURI string, maxProcs int) (err error) {
	if defaultServer, err = NewServer(amqpURI, maxProcs); err == nil {
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

//NewServer constructs a new Server instance which will listen on the provided
//AMQP URI.
func NewServer(amqpURI string, maxProcs int) (ret *Server, err error) {
	ret = &Server{
		uri:           amqpURI,
		maxProcs:      maxProcs,
		reqChan:       make(chan *Request),
		processedChan: make(chan E),
	}
	ret.fileConsumer, err = NewFileConsumer(ret.uri, ret.reqChan)
	if err != nil {
		return
	}
	ret.ender, err = NewEnder(ret.uri, ret.reqChan)
	return
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func (this *Server) Serve() {
	handleFunc := func(h *MessageHandler) {
		herr := h.Handle()
		if herr != nil {
			util.Log(herr)
		}
	}
	go handleFunc(this.fileConsumer)
	go handleFunc(this.ender)
	helpers := make(map[bson.ObjectId]*ProcHelper)
	fileQueues := make(map[bson.ObjectId]*list.List)
	subQueue := list.New()
	busy := 0
	//Begin monitoring processing status
	for {
		if busy < this.maxProcs && subQueue.Len() > 0 {
			//If there is an available spot,
			//start processing the next submission.
			subId := subQueue.Remove(subQueue.Front()).(bson.ObjectId)
			helper := helpers[subId]
			helper.SetStarted()
			if helper.Done() {
				delete(helpers, subId)
			}
			go helper.Handle(this.processedChan, fileQueues[subId])
			delete(fileQueues, subId)
			busy++
		} else if busy < 0 {
			//This will only occur when Shutdown() has been called and
			//all submissions have been completed and processed.
			break
		}
		select {
		case request := <-this.reqChan:
			helper, ok := helpers[request.SubId]
			if request.Stop && ok {
				//If the submission has finished, set the submission's ProcHelper to done
				//and if it has already started, remove it from the queue.
				helper.SetDone()
				if helper.Started() {
					delete(helpers, request.SubId)
				}
			} else if request.Stop {
				util.Log(fmt.Errorf("No submission %q found to end.", request.SubId))
			} else if !ok {
				//This is a new submission so we initialise it.
				subQueue.PushBack(request.SubId)
				helpers[request.SubId] = NewProcHelper(request.SubId)
				fileQueues[request.SubId] = list.New()
				fileQueues[request.SubId].PushBack(request.FileId)
				ChangeStatus(Status{1, 1})
			} else if helper.started {
				//Send file to goroutine if
				//submission processing has started.
				helper.serveChan <- request.FileId
				ChangeStatus(Status{1, 0})
			} else {
				//Add file to queue if not.
				fileQueues[request.SubId].PushBack(request.FileId)
				ChangeStatus(Status{1, 0})
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
	err = this.ender.Shutdown()
	if err != nil {
		return
	}
	err = this.fileConsumer.Shutdown()
	return
}

//NewProcHelper creates a new ProcHelper for the specified
//Submission.
func NewProcHelper(subId bson.ObjectId) *ProcHelper {
	return &ProcHelper{
		subId:     subId,
		serveChan: make(chan bson.ObjectId),
		doneChan:  make(chan E),
		started:   false,
		done:      false,
	}
}

//SetDone indicates that this submission will receive no more files.
func (this *ProcHelper) SetDone() {
	if this.Started() {
		//If it has started send on its channel.
		this.doneChan <- E{}
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
func (this *ProcHelper) Handle(onDone chan E, fileQueue *list.List) {
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
