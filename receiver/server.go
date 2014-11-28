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

//Package receiver provides a TCP server which receives user submissions and
//snapshots and sends them to be processed.
package receiver

import (
	"fmt"

	"github.com/godfried/impendulo/util"

	"net"
	"strconv"
)

type (

	//Spawner is an interface used to spawn Handlers.
	Spawner interface {
		Spawn() Handler
	}

	//Handler is an interface with basic methods for handling connections.
	Handler interface {
		Start(net.Conn)
		End(error)
	}
)

const (
	LOG_SERVER      = "receiver/server.go"
	PORT       uint = 8010
)

//Run is used to listen for new tcp connections and
//spawn a new goroutine for each connection.
//Each goroutine launched will handle its connection and
//its type is determined by HandlerSpawner.
func Run(p uint, s Spawner) {
	//Start listening for connections
	l, e := net.Listen("tcp", ":"+strconv.Itoa(int(p)))
	if e != nil {
		util.Log(fmt.Errorf("error %q listening on port %d", e, p), LOG_SERVER)
		return
	}
	defer l.Close()
	for {
		if c, e := l.Accept(); e != nil {
			util.Log(fmt.Errorf("error %q accepting connection", e), LOG_SERVER)
		} else {
			//Spawn a handler for each new connection.
			go func(cn net.Conn) {
				h := s.Spawn()
				h.Start(cn)
			}(c)
		}
	}
}
