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
		FileCount   int
		Submissions map[string]map[string]util.E
	}

	//Monitor is used to keep track of and change Impendulo's processing
	//status.
	Monitor struct {
		statusChan              chan Status
		requestChan             chan Request
		idleChan                chan util.E
		loader, waiter, changer *MessageHandler
	}
)

var (
	monitor *Monitor
)

func NewStatus() *Status {
	return &Status{FileCount: 0, Submissions: make(map[string]map[string]util.E)}
}

func (s *Status) Update(r Request) error {
	switch r.Type {
	case FILE_ADD:
		return s.addFile(r)
	case FILE_REMOVE:
		return s.removeFile(r)
	case SUBMISSION_START:
		return s.addSubmission(r)
	case SUBMISSION_STOP:
		return s.removeSubmission(r)
	default:
		return fmt.Errorf("unknown request type %d", r.Type)
	}
}

func (s *Status) removeSubmission(r Request) error {
	sk := r.SubId.Hex()
	if fm, ok := s.Submissions[sk]; !ok {
		return fmt.Errorf("submission %s does not exist", sk)
	} else if len(fm) > 0 {
		return fmt.Errorf("submission %s still has active files", sk)
	}
	delete(s.Submissions, sk)
	return nil
}

func (s *Status) addSubmission(r Request) error {
	sk := r.SubId.Hex()
	if _, ok := s.Submissions[sk]; ok {
		return fmt.Errorf("submission %s already exists", sk)
	}
	s.Submissions[sk] = make(map[string]util.E)
	return nil
}

func (s *Status) addFile(r Request) error {
	sk, fk := r.SubId.Hex(), r.FileId.Hex()
	fm, ok := s.Submissions[sk]
	if !ok {
		return fmt.Errorf("submission %s does not exist for file %s", sk, fk)
	}
	if _, ok = fm[fk]; ok {
		return fmt.Errorf("file %s exists for submission %s", fk, sk)
	}
	fm[fk] = util.E{}
	s.FileCount++
	return nil
}

func (s *Status) removeFile(r Request) error {
	sk, fk := r.SubId.Hex(), r.FileId.Hex()
	fm, ok := s.Submissions[sk]
	if !ok {
		return fmt.Errorf("submission %s does not exist for file %s", sk, fk)
	}
	if _, ok = fm[fk]; !ok {
		return fmt.Errorf("file %s does not exist for submission %s", fk, sk)
	}
	delete(fm, fk)
	s.FileCount--
	return nil
}

func (s *Status) Idle() bool {
	return len(s.Submissions) == 0
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
	rc := make(chan Request)
	ic := make(chan util.E)
	c, e := NewChanger(rc)
	if e != nil {
		return nil, e
	}
	w, e := NewWaiter(ic)
	if e != nil {
		return nil, e
	}
	l, e := NewLoader(sc)
	if e != nil {
		return nil, e
	}
	return &Monitor{changer: c, waiter: w, loader: l, statusChan: sc, requestChan: rc, idleChan: ic}, nil
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
	s := NewStatus()
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
