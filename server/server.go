package server

import (
	"errors"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/user"
	"github.com/disco-volante/intlola/proc"
	"github.com/disco-volante/intlola/utils"
	"github.com/disco-volante/intlola/sub"
	"labix.org/v2/mgo/bson"
	"net"
)

const(
 OK = "ok"
 EOF = "eof"
 SEND = "send"
LOGIN = "begin"
 LOGOUT = "end"
 REQ = "req"
 UNAME = "uname"
 PWORD = "pword"
 PROJECT = "project"
 MODE = "mode"
)
type Client struct {
	username string
	project  string
	mode     string
}





/*
Manages an incoming connection request.
*/
func ConnHandler(conn net.Conn, fileChan chan *sub.File) {
	jobj, err := utils.ReadJSON(conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	subId, err := Login(jobj, conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	receiving := true
	for receiving && err == nil{
		jobj, err = utils.ReadJSON(conn)
		if err != nil {
			EndSession(conn, err)
			return
		}	
		req, err := utils.JSONValue(jobj, REQ)
		if err != nil {
			EndSession(conn, err)
			return
		}	
		if req == SEND {
			delete(jobj, REQ)
			err = ProcessFile(subId, jobj, conn, fileChan)
		} else if req == LOGOUT {
			receiving = false
		} else {
			err = errors.New("Unknown request: " + req)
		}
	}
	EndSession(conn, err)
}


/*
Determines whether a file send request is valid and reads a file if it is.
*/
func ProcessFile(subId bson.ObjectId, finfo map[string]interface{}, conn net.Conn, fileChan chan *sub.File) (err error) {
	conn.Write([]byte(OK))
	buffer, err := utils.ReadFile(conn, []byte(EOF))
	if err != nil{
		return err
	}
	conn.Write([]byte(OK))
	fileChan <- sub.NewFile(subId, finfo, buffer.Bytes())
	return nil
}

/*
Determines whether a login request is valid and delivers this 
result to the client 
*/
func Login(jobj map[string]interface{}, conn net.Conn) (subId bson.ObjectId, err error) {
	c, err := createClient(jobj)
	if err == nil {
		s := sub.NewSubmission(c.project, c.username, c.mode)
		err = db.AddSingle(db.SUBMISSIONS, s)
		if err == nil {
			conn.Write([]byte(OK))
			subId = s.Id
		}
	}
	return subId, err
}

/*
Handles an error by logging it as well as reporting it to the connected
user if possible.
*/
func EndSession(conn net.Conn, err error) {
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		utils.Log(msg)
	} else {
		msg = OK
	}
	conn.Write([]byte(msg))
	conn.Close()
}

func createClient(jobj map[string]interface{}) (c *Client, err error) {
	uname, err := utils.JSONValue(jobj, UNAME)
	if err != nil {
		return c, err
	}
	pword, err := utils.JSONValue(jobj, PWORD)
	if err != nil {
		return c, err
	}
	project, err := utils.JSONValue(jobj, PROJECT)
	if err != nil {
		return c, err
	}
	mode, err := utils.JSONValue(jobj, MODE)
	if err != nil {
		return c, err
	}
	umap, err := db.GetById(db.USERS, uname)
	if err == nil {
		usr := user.ReadUser(umap)
		utils.Log(usr, err)
		if usr.CheckSubmit(mode) && utils.Validate(usr.Password, usr.Salt, pword) {
			c = &Client{uname, project, mode}
		} else {
			err = errors.New("Invalid username or password")
		}
	}
	return c, err
}

/*
Listens for new connections and creates a new goroutine for each connection.
*/
func Run(address string, port string) {
	fileChan := make(chan *sub.File)
	go proc.Serve(fileChan)
	service := address + ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		utils.Log("Error: Could not resolve address ", err)
	} else {
		netListen, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if err != nil {
			utils.Log(err)
		} else {
			defer netListen.Close()
			for {
				conn, err := netListen.Accept()
				if err != nil {
					utils.Log("Client error: ", err)
				} else {
					go ConnHandler(conn, fileChan)
				}
			}
		}
	}
}
