package server

import (
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/processing"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"github.com/godfried/cabanga/user"
	"labix.org/v2/mgo/bson"
	"net"
	"runtime"
	"fmt"
)

const (
	OK      = "ok"
	SEND    = "send"
	LOGIN   = "begin"
	LOGOUT  = "end"
	REQ     = "req"
)

/*
Listens for new connections and creates a new goroutine for each connection.
*/
func Run(port string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	go processing.Serve(subChan, fileChan)
	netListen, err := net.Listen("tcp", ":" + port)
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
			go connHandler(conn, subChan, fileChan)
		}
	}
}


/*
Manage incoming connection request.
*/
func connHandler(conn net.Conn, subChan chan *submission.Submission, fileChan chan *submission.File) {
	jobj, err := util.ReadJSON(conn)
	if err != nil {
		endSession(conn, err)
		return
	}
	sub, err := login(jobj, conn)
	if err != nil {
		endSession(conn, err)
		return
	}
	subChan <- sub
	util.Log("Created submission: ", sub)
	receiving := true
	for receiving && err == nil {
		jobj, err = util.ReadJSON(conn)
		if err != nil {
			util.Log(err)
			endSession(conn, err)
			return
		}
		req, err := util.GetString(jobj, REQ)
		if err != nil {
			util.Log(err)
			endSession(conn, err)
			return
		}
		if req == SEND {
			delete(jobj, REQ)
			err = processFile(sub.Id, jobj, conn, fileChan)
		} else if req == LOGOUT {
			receiving = false
			util.Log("Completed submission: ", sub)
		} else {
			err = fmt.Errorf("Unknown request %q", req)
		}
	}
	endSession(conn, err)
}

/*
Reads file data from connection and sends data to be processed.
*/
func processFile(subId bson.ObjectId, finfo map[string]interface{}, conn net.Conn, fileChan chan *submission.File) error {
	conn.Write([]byte(OK))
	buffer, err := util.ReadData(conn)
	if err != nil {
		return err
	}
	conn.Write([]byte(OK))
	f := submission.NewFile(subId, finfo, buffer)
	err = db.AddFile(f)
	if err != nil {
		return err
	}
	fileChan <- f
	return nil
}

/*
Creates a new submission if the login request is valid.
*/
func login(jobj map[string]interface{}, conn net.Conn) (*submission.Submission, error) {
	sub, err := createSubmission(jobj)
	if err != nil {
		return nil, err
	}
	err = db.AddSubmission(sub)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(OK))
	return sub, nil
}

/*
Ends a client session and reports any errors to the client. 
*/
func endSession(conn net.Conn, err error) {
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		util.Log(err)
	} else {
		msg = OK
	}
	conn.Write([]byte(msg))
	conn.Close()
}

/*
Reads submission values from a json object.
Determines whether user has neccesary privileges for submission and is using correct password 
*/
func createSubmission(jobj map[string]interface{}) (*submission.Submission, error) {
	username, err := util.GetString(jobj, submission.USER)
	if err != nil {
		return nil, err
	}
	pword, err := util.GetString(jobj, user.PWORD)
	if err != nil {
		return nil, err
	}
	project, err := util.GetString(jobj, submission.PROJECT)
	if err != nil {
		return nil, err
	}
	mode, err := util.GetString(jobj, submission.MODE)
	if err != nil {
		return nil, err
	}
	lang, err := util.GetString(jobj, submission.LANG)
	if err != nil {
		return nil, err
	}
	u, err := db.GetUserById(username)
	if err != nil {
		return nil, err
	}
	if !u.CheckSubmit(mode){
		return nil, fmt.Errorf("User %q has insufficient permissions for %q", username, mode)
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		return nil, fmt.Errorf("User %q attempted to login with an invalid username or password", username)
	}
	return submission.NewSubmission(project, username, mode, lang), nil
}


