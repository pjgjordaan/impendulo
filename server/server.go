//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

//Package server provides a TCP server which receives user submissions and
//snapshots and sends them to be processed.
package server

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"net"
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

//Run is used to listen for new tcp connections and
//spawn a new goroutine for each connection.
//Each goroutine launched will handle its connection and
//its type is determined by HandlerSpawner.
func Run(port string, spawner HandlerSpawner) {
	//Start listening for connections
	netListen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		util.Log(fmt.Errorf(
			"Encountered error %q when listening on port %q",
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
