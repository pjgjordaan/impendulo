package server

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"io"
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
const LOG_SERVER = "server/server.go"

/*type connStruct struct{
	conn net.Conn
	add bool
}

var connChan chan connStruct
var allowChan chan bool

func manageConnections(){
	connChan = make(chan connStruct)
	allowChan = make(chan bool)
	connected := make(map[string]bool)
	for{
		conn, ok := <- connChan
		fmt.Println(conn)
		if !ok{
			break
		}
		host, _, err := net.SplitHostPort(conn.conn.LocalAddr().String())
		if conn.add{
			if err != nil || connected[host]{
				allowChan <- false
			} else{
				connected[host] = true
				allowChan <- true
			}
		} else{
			delete(connected, host)
		}
	}
}*/

//Run is used to listen for new tcp connections and
//spawn a new goroutine for each connection.
//Each goroutine launched will handle its connection and
//its type is determined by HandlerSpawner.
func Run(port string, spawner HandlerSpawner) {
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
			go func(c net.Conn){
				handler := spawner.Spawn()
				handler.Start(c)
			}(conn)
		}
	}
}

//HandlerSpawner is an interface used to spawn RWCHandlers.
type HandlerSpawner interface {
	Spawn() RWCHandler
}

//ConnHandler is an interface with basic methods for handling connections.
type RWCHandler interface {
	Start(rwc io.ReadWriteCloser)
	End(err error)
}
