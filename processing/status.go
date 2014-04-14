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
	"fmt"

	"github.com/godfried/impendulo/util"
)

type (

	//Status is used to indicate a change in the files or
	//submissions being processed. It is also used to retrieve the current
	//number of files and submissions being processed.
	Status struct {
		Files       int
		Submissions int
	}

	//Monitor is used to keep track of and change Impendulo's processing
	//status.
	Monitor struct {
		statusChan              chan Status
		loader, waiter, changer *MessageHandler
	}
)

var (
	monitor *Monitor
)

//add adds the value of toAdd to this Status.
func (s *Status) add(toAdd Status) {
	s.Files += toAdd.Files
	s.Submissions += toAdd.Submissions
}

//MonitorStatus begins keeping track of Impendulo's current processing status.
func MonitorStatus() error {
	var e error
	if monitor != nil {
		if e = monitor.Shutdown(); e != nil {
			return e
		}
	}
	if monitor, e = NewMonitor(); e != nil {
		return e
	}
	go monitor.Monitor()
	return nil
}

//NewMonitor
func NewMonitor() (*Monitor, error) {
	sc := make(chan Status)
	c, e := NewChanger(sc)
	if e != nil {
		return nil, e
	}
	w, e := NewWaiter(sc)
	if e != nil {
		return nil, e
	}
	l, e := NewLoader(sc)
	if e != nil {
		return nil, e
	}
	return &Monitor{changer: c, waiter: w, loader: l, statusChan: sc}, nil
}

//Monitor begins a new monitoring session for this Monitor.
//It handles status updates and requests.
func (m *Monitor) Monitor() {
	h := func(mh *MessageHandler) {
		if e := mh.Handle(); e != nil {
			util.Log(e, mh.Shutdown())
		}
	}
	go h(m.changer)
	go h(m.loader)
	go h(m.waiter)
	s := new(Status)
	for v := range m.statusChan {
		switch v {
		case Status{}:
			//A zeroed Status indicates a request for the current Status.
			m.statusChan <- *s
		default:
			s.add(v)
			util.Log(fmt.Sprintf("Status updated with %v to %v.", v, *s))
		}
	}
}

func ShutdownMonitor() error {
	if monitor == nil {
		return nil
	}
	if e := monitor.Shutdown(); e != nil {
		return e
	}
	monitor = nil
	return nil
}

//Shutdown stops this Monitor
func (m *Monitor) Shutdown() error {
	close(m.statusChan)
	return m.shutdownHandlers()
}

//shutdownHandlers stops all the MesageHandlers used by this Monitor.
func (m *Monitor) shutdownHandlers() error {
	if e := m.waiter.Shutdown(); e != nil {
		return e
	}
	if e := m.loader.Shutdown(); e != nil {
		return e
	}
	return m.changer.Shutdown()
}
