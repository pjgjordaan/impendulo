package server

import (
	"errors"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/processing"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"net"
"runtime"
)

type Client struct {
	username string
	project  string
	mode     string
	lang     string
}

/*
Manage incoming connection request.
*/
func ConnHandler(conn net.Conn, fileChan chan processing.Item) {
	jobj, err := utils.ReadJSON(conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	subId, err := Login(jobj, conn)
	utils.Log(subId)
	if err != nil {
		EndSession(conn, err)
		return
	}
	utils.Log("Created submission: ", subId)
	receiving := true
	for receiving && err == nil {
		utils.Log("w1", jobj)
		jobj, err = utils.ReadJSON(conn)
		if err != nil {
			utils.Log("JSON error: ", err)
			EndSession(conn, err)
			return
		}
		utils.Log("w2", jobj)
		req, err := utils.GetString(jobj, REQ)
		if err != nil {
			utils.Log("JSON error: ", err)
			EndSession(conn, err)
			return
		}
		utils.Log("w1",req, jobj)
		if req == SEND {
			delete(jobj, REQ)
			err = ProcessFile(subId, jobj, conn, fileChan)
		} else if req == LOGOUT {
			receiving = false
			utils.Log("Completed submission: ", subId)
		} else {
			err = errors.New("Unknown request: " + req)
		}
		utils.Log("received", jobj)
	}
	EndSession(conn, err)
}

/*
Reads file data from connection and sends data to be processed.
*/
func ProcessFile(subId bson.ObjectId, finfo map[string]interface{}, conn net.Conn, fileChan chan processing.Item) error {
	conn.Write([]byte(OK))
	buffer, err := utils.ReadData(conn)
	if err != nil {
		utils.Log("Conn read error: ", err)
		return err
	}
	utils.Log("Read file: ", finfo)
	conn.Write([]byte(OK))
	f := submission.NewFile(subId, finfo, buffer)
	err = db.AddFile(f)
	if err != nil {
		utils.Log("DB error: ", err)
		return err
	}
	utils.Log("Saved file: ", f.Id)
	fileChan <- processing.Item{f.Id, f.SubId}
	utils.Log("Sent file: ", f.Id)
	return nil
}

/*
Creates a new submission if the login request is valid.
*/
func Login(jobj map[string]interface{}, conn net.Conn) (subId bson.ObjectId, err error) {
	c, err := createClient(jobj)
	if err != nil {
		utils.Log("Login error: ", err)
		return subId, err
	}
	s := submission.NewSubmission(c.project, c.username, c.mode, c.lang)
	err = db.AddSubmission(s)
	if err != nil {
		utils.Log("DB error: ", err)
		return subId, err
	}
	conn.Write([]byte(OK))
	subId = s.Id
	return subId, err
}

/*
Ends a client session and reports any errors to the client. 
*/
func EndSession(conn net.Conn, err error) {
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		utils.Log("Sending error: ", msg)
	} else {
		msg = OK
	}
	conn.Write([]byte(msg))
	conn.Close()
}

/*
Reads client values from a json object.
Determines whether client has neccesary privileges for submission and is using correct password 
*/
func createClient(jobj map[string]interface{}) (*Client, error) {
	uname, err := utils.GetString(jobj, UNAME)
	if err != nil {
		return nil, err
	}
	pword, err := utils.GetString(jobj, PWORD)
	if err != nil {
		return nil, err
	}
	project, err := utils.GetString(jobj, PROJECT)
	if err != nil {
		return nil, err
	}
	mode, err := utils.GetString(jobj, MODE)
	if err != nil {
		return nil, err
	}
	lang, err := utils.GetString(jobj, LANG)
	if err != nil {
		return nil, err
	}
	u, err := db.GetUserById(uname)
	if err != nil {
		return nil, err
	}
	if !u.CheckSubmit(mode) || !utils.Validate(u.Password, u.Salt, pword) {
		return nil, errors.New("Invalid username or password")
	}
	return &Client{uname, project, mode, lang}, nil
}

/*
Listens for new connections and creates a new goroutine for each connection.
*/
func Run(port string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fileChan := make(chan processing.Item)
	go processing.Serve(fileChan)
	netListen, err := net.Listen("tcp", ":" + port)
	if err != nil {
		utils.Log("Listening error: ", err)
		return
	}
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

const (
	OK      = "ok"
	SEND    = "send"
	LOGIN   = "begin"
	LOGOUT  = "end"
	REQ     = "req"
	UNAME   = "uname"
	PWORD   = "pword"
	PROJECT = "project"
	MODE    = "mode"
	LANG    = "lang"
)
