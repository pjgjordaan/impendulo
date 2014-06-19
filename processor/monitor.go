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

package processor

import (
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/status"
	"github.com/godfried/impendulo/util"
)

type (

	//Monitor is used to keep track of and change Impendulo's processing
	//status.
	Monitor struct {
		statusChan              chan status.S
		requestChan             chan *request.R
		idleChan                chan util.E
		loader, waiter, changer *mq.MessageHandler
	}
)

var (
	monitor *Monitor
)

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
	sc := make(chan status.S)
	rc := make(chan *request.R)
	ic := make(chan util.E)
	c, e := mq.NewChanger(rc)
	if e != nil {
		return nil, e
	}
	w, e := mq.NewWaiter(ic)
	if e != nil {
		return nil, e
	}
	l, e := mq.NewLoader(sc)
	if e != nil {
		return nil, e
	}
	return &Monitor{changer: c, waiter: w, loader: l, statusChan: sc, requestChan: rc, idleChan: ic}, nil
}

//Monitor begins a new monitoring session for this Monitor.
//It handles status updates and requests.
func (m *Monitor) Monitor() {
	go mq.H(m.changer)
	go mq.H(m.loader)
	go mq.H(m.waiter)
	s := status.New()
	waiting := 0
	for {
		if waiting > 0 && s.Idle() {
			if e := m.notifyWaiting(waiting); e != nil {
				util.Log(e)
				return
			}
			waiting = 0
		}
		select {
		case _, ok := <-m.statusChan:
			if !ok {
				return
			}
			m.statusChan <- *s
		case r, ok := <-m.requestChan:
			if !ok {
				return
			}
			if e := s.Update(r); e != nil {
				util.Log(e)
			}
		case <-m.idleChan:
			waiting++
		}
	}
}

func (m *Monitor) notifyWaiting(n int) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
	}()
	for i := 0; i < n; i++ {
		m.idleChan <- util.E{}
	}
	return
}

//ShutdownMonitor
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
	close(m.requestChan)
	close(m.idleChan)
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
