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
func (this *Status) add(toAdd Status) {
	this.Files += toAdd.Files
	this.Submissions += toAdd.Submissions
}

//MonitorStatus begins keeping track of Impendulo's current processing status.
func MonitorStatus() (err error) {
	if monitor != nil {
		err = monitor.Shutdown()
		if err != nil {
			return
		}
	}
	monitor, err = NewMonitor()
	if err != nil {
		return
	}
	go monitor.Monitor()
	return
}

//NewMonitor
func NewMonitor() (ret *Monitor, err error) {
	ret = &Monitor{
		statusChan: make(chan Status),
	}
	ret.changer, err = NewChanger(ret.statusChan)
	if err != nil {
		return
	}
	ret.waiter, err = NewWaiter(ret.statusChan)
	if err != nil {
		return
	}
	ret.loader, err = NewLoader(ret.statusChan)
	return
}

//Monitor begins a new monitoring session for this Monitor.
//It handles status updates and requests.
func (this *Monitor) Monitor() {
	handle := func(mh *MessageHandler) {
		merr := mh.Handle()
		if merr != nil {
			util.Log(merr, mh.Shutdown())
		}
	}
	go handle(this.changer)
	go handle(this.loader)
	go handle(this.waiter)
	var status *Status = new(Status)
	for val := range this.statusChan {
		switch val {
		case Status{}:
			//A zeroed Status indicates a request for the current Status.
			this.statusChan <- *status
		default:
			status.add(val)
			util.Log(fmt.Sprintf("Status updated with %v to %v.", val, *status))
		}
	}
}

func ShutdownMonitor() (err error) {
	if monitor == nil {
		return
	}
	err = monitor.Shutdown()
	if err == nil {
		monitor = nil
	}
	return
}

//Shutdown stops this Monitor
func (this *Monitor) Shutdown() (err error) {
	close(this.statusChan)
	err = this.shutdownHandlers()
	return
}

//shutdownHandlers stops all the MesageHandlers used by this Monitor.
func (this *Monitor) shutdownHandlers() (err error) {
	err = this.waiter.Shutdown()
	if err != nil {
		return
	}
	err = this.loader.Shutdown()
	if err != nil {
		return
	}
	err = this.changer.Shutdown()
	return
}