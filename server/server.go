package server

import (
	"errors"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/usr"
	"github.com/disco-volante/intlola/proc"
	"github.com/disco-volante/intlola/utils"
	"github.com/disco-volante/intlola/submission"
	"labix.org/v2/mgo/bson"
	"net"
)

const FNAME = "FILENAME"
const FTYPE = "FILETYPE"
const OK = "OK"
const EOF = "EOF"
const SEND = "SEND"
const LOGOUT = "LOGOUT"
const REQ = "TYPE"
const UNAME = "USERNAME"
const PWORD = "PASSWORD"
const PROJECT = "PROJECT"
const MODE = "MODE"

type Client struct {
	username string
	project  string
	mode     string
}

/*
Determines whether a file send request is valid and reads a file if it is.
*/
func ProcessFile(subId bson.ObjectId, jobj map[string]interface{}, conn net.Conn, fileChan chan *submission.File) (err error) {
	fname, err := utils.JSONValue(jobj, FNAME)
	if err == nil {
		ftype, err := utils.JSONValue(jobj, FTYPE)
		if err == nil {
			conn.Write([]byte(OK))
			buffer, err := utils.ReadFile(conn, []byte(EOF))
			if err == nil {
				f := submission.NewFile(subId, fname, ftype, buffer.Bytes())
				fileChan <- f
				conn.Write([]byte(OK))
			}
		}
	}
	return err
}

/*
Manages an incoming connection request.
*/
func ConnHandler(conn net.Conn, fileChan chan *submission.File) {
	jobj, err := utils.ReadJSON(conn)
	if err == nil {
		subId, err := Login(jobj, conn)
		for err == nil {
			jobj, err = utils.ReadJSON(conn)
			if err == nil {
				req, err := utils.JSONValue(jobj, REQ)
				if req == SEND {
					err = ProcessFile(subId, jobj, conn, fileChan)
				} else if req == LOGOUT {
					break
				} else if err == nil {
					err = errors.New("Unknown request: " + req)
				}
			}
		}
	}
	EndSession(conn, err)
}

/*
Determines whether a login request is valid and delivers this 
result to the client 
*/
func Login(jobj map[string]interface{}, conn net.Conn) (subId bson.ObjectId, err error) {
	c, err := createClient(jobj)
	if err == nil {
		sub := submission.NewSubmission(c.project, c.username, c.mode)
		err = db.AddSingle(db.SUBMISSIONS, sub)
		if err == nil {
			conn.Write([]byte(OK))
			subId = sub.Id
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
	uint, err := db.GetById(uname, db.USERS)
	if err == nil {	
		user := uint.(*usr.User)
		if user.CheckSubmit(mode) && utils.Validate(user.Password, user.Salt, pword) {
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
	fileChan := make(chan *submission.File)
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
