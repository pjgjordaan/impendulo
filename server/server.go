package server

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"net"
)

const (
	OK     = "ok"
	SEND   = "send"
	LOGIN  = "begin"
	LOGOUT = "end"
	REQ    = "req"
)

//Run listens for new connections and creates a new goroutine for each connection.
func RunFileReceiver(port string, subChan chan *submission.Submission, fileChan chan *submission.File) {
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
			go ConnHandler(conn, subChan, fileChan)
		}
	}
}

func RunTestReceiver(port string){
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
			fmt.Println("received connection")
			go ReceiveTests(conn)
		}
	}
}

//ConnHandler manages an incoming connection request.
//It authenticates the request and processes files sent on the connection.
func ReceiveTests(conn net.Conn) {
	testInfo, err := util.ReadJSON(conn)
	fmt.Println("received info", testInfo)
	if err != nil {
		EndSession(conn, err)
		return
	}
	project, err := util.GetString(testInfo, submission.PROJECT)
	if err != nil {
		EndSession(conn, err)
		return
	}
	lang, err := util.GetString(testInfo, submission.LANG)
	if err != nil {
		EndSession(conn, err)
		return
	}
	names, err := util.GetStrings(testInfo, submission.NAMES)
	if err != nil {
		EndSession(conn, err)
		return
	}
	conn.Write([]byte(OK))
	testFiles, err := util.ReadData(conn)
	if err != nil {
		EndSession(conn, err)
		return 
	}
	fmt.Println("received tests", testFiles)
	conn.Write([]byte(OK))
	dataFiles, err := util.ReadData(conn)
	if err != nil {
		EndSession(conn, err)
		return 
	}
	fmt.Println("received tests", dataFiles)
	conn.Write([]byte(OK))
	test := submission.NewTest(project, lang, names, testFiles, dataFiles)
	err = db.AddTest(test)
	EndSession(conn, err)
}

//ConnHandler manages an incoming connection request.
//It authenticates the request and processes files sent on the connection.
func ConnHandler(conn net.Conn, subChan chan *submission.Submission, fileChan chan *submission.File) {
	jobj, err := util.ReadJSON(conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	sub, err := Login(jobj, conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	subChan <- sub
	util.Log("Created submission: ", sub)
	receiving := true
	for receiving && err == nil {
		jobj, err = util.ReadJSON(conn)
		if err != nil {
			EndSession(conn, err)
			return
		}
		req, err := util.GetString(jobj, REQ)
		if err != nil {
			EndSession(conn, err)
			return
		}
		if req == SEND {
			delete(jobj, REQ)
			err = ProcessFile(sub.Id, jobj, conn, fileChan)
		} else if req == LOGOUT {
			receiving = false
		} else {
			err = fmt.Errorf("Unknown request %q", req)
		}
	}
	EndSession(conn, err)
}

//processFile reads file data from connection and stores it in the db.
//The file data is then sent on fileChan for further processing.
func ProcessFile(subId bson.ObjectId, finfo map[string]interface{}, conn net.Conn, fileChan chan *submission.File) error {
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

//Login creates a new submission if the login request is valid.
func Login(jobj map[string]interface{}, conn net.Conn) (*submission.Submission, error) {
	sub, err := CreateSubmission(jobj)
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

//EndSession ends a session and reports any errors to the client. 
func EndSession(conn net.Conn, err error) {
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

//CreateSubmission validates a login request. 
//It reads submission values from a json object and checks privilege level and password. 
func CreateSubmission(jobj map[string]interface{}) (*submission.Submission, error) {
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
	if !u.CheckSubmit(mode) {
		return nil, fmt.Errorf("User %q has insufficient permissions for %q", username, mode)
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		return nil, fmt.Errorf("User %q attempted to login with an invalid username or password", username)
	}
	return submission.NewSubmission(project, username, mode, lang), nil
}
