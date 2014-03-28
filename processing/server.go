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

	//Helper is used to help handle a submission's files.
	Helper struct {
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
func Serve(maxProcs uint) error {
	var e error
	if defaultServer, e = NewServer(maxProcs); e != nil {
		return e
	}
	defaultServer.Serve()
	return nil
}

//Shutdown signals to the default Server that it can shutdown
//and waits for it to complete all processing. It then shuts down all
//active producers as well as the status monitor.
func Shutdown() error {
	if e := defaultServer.Shutdown(); e != nil {
		return e
	}
	if e := StopProducers(); e != nil {
		return e
	}
	return ShutdownMonitor()
}

//NewServer constructs a new Server instance which will listen on the coinfigured
//AMQP URI.
func NewServer(maxProcs uint) (*Server, error) {
	rc := make(chan *Request)
	sm, st, e := NewSubmitter(rc)
	if e != nil {
		return nil, e
	}
	r, e := NewRedoer(rc)
	if e != nil {
		return nil, e
	}
	return &Server{
		maxProcs:      maxProcs,
		requestChan:   rc,
		processedChan: make(chan E),
		submitter:     sm,
		starter:       st,
		redoer:        r,
	}, nil
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func (s *Server) Serve() {
	go handleFunc(s.starter)
	go handleFunc(s.submitter)
	go handleFunc(s.redoer)
	hm := make(map[bson.ObjectId]*Helper)
	fq := make(map[bson.ObjectId]*list.List)
	sq := list.New()
	var busy uint = 0
	//Begin monitoring processing status
	for {
		if busy < s.maxProcs && sq.Len() > 0 {
			//If there is an available spot,
			//start processing the next submission.
			sid := sq.Remove(sq.Front()).(bson.ObjectId)
			h := hm[sid]
			h.started = true
			if h.done {
				delete(hm, sid)
			}
			go h.Handle(s.processedChan, fq[sid])
			delete(fq, sid)
			busy++
		} else if busy < 0 {
			//This will only occur when Shutdown() has been called and
			//all submissions have been completed and processed.
			break
		}
		select {
		case r := <-s.requestChan:
			h, ok := hm[r.SubId]
			switch r.Type {
			case SUBMISSION_STOP:
				if !ok {
					util.Log(fmt.Errorf("no submission %q found to end", r.SubId))
				} else {
					//If the submission has finished, set the submission's Helper to done
					//and if it has already started, remove it from the queue.
					h.SetDone()
					if h.started {
						delete(hm, r.SubId)
					}
				}
			case SUBMISSION_START:
				if ok {
					util.Log(fmt.Errorf("submission %s already started", r.SubId))
				} else {
					//This is a new submission so we initialise it.
					sq.PushBack(r.SubId)
					hm[r.SubId] = NewHelper(r.SubId)
					fq[r.SubId] = list.New()
					ChangeStatus(Status{Files: 0, Submissions: 1})
				}
			case FILE_ADD:
				if !ok {
					util.Log(fmt.Errorf("no submission %s found for file %", r.SubId, r.FileId))
				} else {
					if h.started {
						//Send file to goroutine if
						//submission processing has started.
						h.serveChan <- r.FileId
					} else {
						//Add file to queue if not.
						fq[r.SubId].PushBack(r.FileId)
					}
					ChangeStatus(Status{Files: 1, Submissions: 0})
				}
			default:
				util.Log(fmt.Errorf("unsupported request type %d", r.Type))
			}
		case <-s.processedChan:
			//A submission has been processed so one less goroutine to worry about.
			busy--
		}
	}
}

//Shutdown stops Serve from running once all submissions have been processed.
func (s *Server) Shutdown() error {
	s.processedChan <- E{}
	if e := s.submitter.Shutdown(); e != nil {
		return e
	}
	if e := s.starter.Shutdown(); e != nil {
		return e
	}
	return s.redoer.Shutdown()
}

//NewHelper creates a new Helper for the specified
//Submission.
func NewHelper(sid bson.ObjectId) *Helper {
	return &Helper{
		subId:     sid,
		serveChan: make(chan bson.ObjectId),
		doneChan:  make(chan E),
		started:   false,
		done:      false,
	}
}

//SetDone indicates that this submission will receive no more files.
func (h *Helper) SetDone() {
	if h.started {
		//If it has started send on its channel.
		h.doneChan <- E{}
	} else {
		//Otherwise simply set done to true.
		h.done = true
	}
}

//Handle helps manage the files a submission receives.
//It spawns a new Processor which runs in a seperate goroutine
//and receives files to process from this Helper.
//fq is the queue of files the submission has received
//prior to the start of processing.
func (h *Helper) Handle(onDone chan E, fq *list.List) {
	pc := make(chan bson.ObjectId)
	sc := make(chan E)
	p, e := NewProcessor(h.subId)
	if e != nil {
		util.Log(e, LOG_SERVER)
		ChangeStatus(Status{-fq.Len(), -1})
		onDone <- E{}
		return
	}
	go p.Process(pc, sc)
	busy := false
	for {
		if !busy {
			if fq.Len() > 0 {
				//Not busy and there are files so send one to be processed.
				fid := fq.Remove(fq.Front()).(bson.ObjectId)
				pc <- fid
				busy = true
			} else if h.done {
				//Not busy and we are done so we should finish up here.
				sc <- E{}
				<-sc
				ChangeStatus(Status{0, -1})
				onDone <- E{}
				return
			}
		}
		select {
		case fid := <-h.serveChan:
			//Add new files to the queue.
			fq.PushBack(fid)
		case <-pc:
			//Processor has finished with its current file.
			busy = false
			ChangeStatus(Status{-1, 0})
		case <-h.doneChan:
			//Submission will receive no more files.
			h.done = true
		}
	}
}
