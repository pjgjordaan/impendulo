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
package processor

import (
	"fmt"

	"github.com/godfried/impendulo/processor/monitor"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/worker/file"
	"github.com/godfried/impendulo/util"

	"runtime"
)

type (

	//Server is our processing server which receives and processes submissions and files.
	Server struct {
		maxProcs    int
		requestChan chan *request.R
		createdChan chan util.E
		submitter   *mq.MessageHandler
	}
)

const (
	LOG_SERVER = "processing/server.go"
)

var (
	defaultServer *Server
	MAX_PROCS     = util.Max(runtime.NumCPU()-1, 1)
)

//Serve launches the default Server. It listens on the configured AMQP URI and
//spawns at most maxProcs goroutines in order to process submissions.
func Serve(maxProcs int) error {
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
	if e := mq.ShutdownProducers(); e != nil {
		return e
	}
	return monitor.Shutdown()
}

//NewServer constructs a new Server instance which will listen on the coinfigured
//AMQP URI.
func NewServer(maxProcs int) (*Server, error) {
	rc := make(chan *request.R)
	s, e := mq.NewSubmitter(rc)
	if e != nil {
		return nil, e
	}
	return &Server{
		maxProcs:    maxProcs,
		requestChan: rc,
		submitter:   s,
	}, nil
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func (s *Server) Serve() {
	go mq.H(s.submitter)
	busy := 0
	processedChan := make(chan util.E, s.maxProcs)
	for {
		if busy < 0 {
			break
		}
		for busy >= s.maxProcs {
			<-processedChan
			busy--
		}
		select {
		case r := <-s.requestChan:
			switch r.Type {
			case request.SUBMISSION_START:
				w, e := file.New(r.SubId)
				if e != nil {
					util.Log(e)
				} else {
					go w.Start(processedChan)
					busy++
				}
			default:
				util.Log(fmt.Errorf("unsupported request type %d", r.Type))
			}
		case <-processedChan:
			busy--
		}
	}
}

//Shutdown stops Serve from running once all submissions have been processed.
func (s *Server) Shutdown() error {
	if e := mq.WaitIdle(); e != nil {
		return e
	}
	return s.submitter.Shutdown()
}
