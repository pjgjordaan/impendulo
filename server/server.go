package server

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"net"
)

const (
	OK         = "ok"
	SEND       = "send"
	LOGIN      = "begin"
	LOGOUT     = "end"
	REQ        = "req"
	PROJECTS   = "projects"
	SUBMISSION = "submission"
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

//HandlerSpawner is an interface used to spawn new ConnHandlers.
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

type BasicSpawner struct{}

func (this *BasicSpawner) Spawn() ConnHandler {
	return new(BasicHandler)
}

type BasicHandler struct {
	Conn net.Conn
}

func (this *BasicHandler) Start(conn net.Conn) {
	this.Conn = conn
	this.End(this.Handle())
}

func (this *BasicHandler) Handle() error {
	err := this.Login()
	if err != nil {
		return err
	}
	return this.Read()

}

func (this *BasicHandler) Login() error {
	_, err := util.ReadJSON(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	return nil
}

func (this *BasicHandler) LoadInfo() error {
	_, err := util.ReadJSON(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	return nil
}

func (this *BasicHandler) Read() error {
	_, err := util.ReadData(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	return nil
}

func (this *BasicHandler) End(err error) {
	defer this.Conn.Close()
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		util.Log(err)
	} else {
		msg = OK
	}
	this.Conn.Write([]byte(msg))
}
