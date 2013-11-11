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

	//HandlerSpawner is an interface used to spawn ConnHandlers.
	HandlerSpawner interface {
		Spawn() ConnHandler
	}

	//ConnHandler is an interface with basic methods for handling connections.
	ConnHandler interface {
		Start(net.Conn)
		End(error)
	}
)

const (
	LOG_SERVER = "receiver/server.go"
)

//Run is used to listen for new tcp connections and
//spawn a new goroutine for each connection.
//Each goroutine launched will handle its connection and
//its type is determined by HandlerSpawner.
func Run(port int, spawner HandlerSpawner) {
	//Start listening for connections
	netListen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		util.Log(fmt.Errorf(
			"Encountered error %q when listening on port %d",
			err, port), LOG_SERVER)
		return
	}
	defer netListen.Close()
	for {
		conn, err := netListen.Accept()
		if err != nil {
			util.Log(fmt.Errorf(
				"Encountered error %q when accepting connection",
				err), LOG_SERVER)
		} else {
			//Spawn a handler for each new connection.
			go func(c net.Conn) {
				handler := spawner.Spawn()
				handler.Start(c)
			}(conn)
		}
	}
}
