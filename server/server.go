package server

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"net"
)

const (
	OK                  = "ok"
	SEND                = "send"
	LOGIN               = "begin"
	LOGOUT              = "end"
	REQ                 = "req"
	PROJECTS            = "projects"
	SUBMISSION_NEW      = "submission_new"
	SUBMISSION_CONTINUE = "submission_continue"
)

//Run is used to listen for new tcp connections and spawn a new goroutine for each connection.
//Each goroutine launched will handle its connection and its type is determined by HandlerSpawner.
func Run(port string, spawner HandlerSpawner) {
	netListen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		util.Log(fmt.Errorf("Encountered error %q when listening on port %q", err, port))
		return
	}
	defer netListen.Close()
	for {
		conn, err := netListen.Accept()
		if err != nil {
			util.Log(fmt.Errorf("Encountered error %q when accepting connection", err))
		} else {
			handler := spawner.Spawn()
			go handler.Start(conn)
		}
	}
}

type HandlerSpawner interface {
	Spawn() ConnHandler
}

//ConnHandler is an interface with basic methods for handling connections.
type ConnHandler interface {
	Start(conn net.Conn)
	Handle() error
	Login() error
	LoadInfo() error
	Read() error
	End(err error)
}