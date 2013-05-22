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
	err := TestLogin(conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	test, err := ReadTest(conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	err = db.AddTest(test)
	EndSession(conn, err)
}

//Login creates a new submission if the login request is valid.
func TestLogin(conn net.Conn) error {
	loginInfo, err := util.ReadJSON(conn)
	if err != nil {
		return err
	}
	username, err := util.GetString(loginInfo, submission.USER)
	if err != nil {
		return err
	}
	pword, err := util.GetString(loginInfo, user.PWORD)
	if err != nil {
		return err
	}
	u, err := db.GetUserById(username)
	if err != nil {
		return err
	}
	if !u.CheckSubmit(submission.TEST_MODE) {
		return fmt.Errorf("User %q has insufficient permissions for %q", username, submission.TEST_MODE)
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		return fmt.Errorf("User %q attempted to login with an invalid username or password", username)
	}
	conn.Write([]byte(OK))
	return nil
}

func ReadTest(conn net.Conn) (*submission.Test, error){
	testInfo, err := util.ReadJSON(conn)
	if err != nil {
		return nil, err
	}
	project, err := util.GetString(testInfo, submission.PROJECT)
	if err != nil {
		return nil, err
	}
	lang, err := util.GetString(testInfo, submission.LANG)
	if err != nil {
		return nil, err
	}
	names, err := util.GetStrings(testInfo, submission.NAMES)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(OK))
	testFiles, err := util.ReadData(conn)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(OK))
	dataFiles, err := util.ReadData(conn)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(OK))
	return submission.NewTest(project, lang, names, testFiles, dataFiles), nil
}

//ConnHandler manages an incoming connection request.
//It authenticates the request and processes files sent on the connection.
func ConnHandler(conn net.Conn, subChan chan *submission.Submission, fileChan chan *submission.File) {
	err := ReceiveFiles(conn, subChan, fileChan)
	EndSession(conn, err)
}

func ReceiveFiles(conn net.Conn, subChan chan *submission.Submission, fileChan chan *submission.File)error{
	sub, err := Login(conn)
	if err != nil {
		return err
	}
	subChan <- sub
	for {
		requestInfo, err := util.ReadJSON(conn)
		if err != nil {
			return err
		}
		req, err := util.GetString(requestInfo, REQ)
		if err != nil {
			return err
		}
		if req == SEND {
			err = ProcessFile(sub.Id, requestInfo, conn, fileChan)
			if err != nil {
				return err
			}
		} else if req == LOGOUT {
			return nil
		} else{
			return fmt.Errorf("Unknown request %q", req)
		}
	} 
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
func Login(conn net.Conn) (*submission.Submission, error) {
	fileInfo, err := util.ReadJSON(conn)
	if err != nil {
		return nil, err
	}
	sub, err := CreateSubmission(fileInfo)
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
func CreateSubmission(loginInfo map[string]interface{}) (*submission.Submission, error) {
	username, err := util.GetString(loginInfo, submission.USER)
	if err != nil {
		return nil, err
	}
	pword, err := util.GetString(loginInfo, user.PWORD)
	if err != nil {
		return nil, err
	}
	project, err := util.GetString(loginInfo, submission.PROJECT)
	if err != nil {
		return nil, err
	}
	mode, err := util.GetString(loginInfo, submission.MODE)
	if err != nil {
		return nil, err
	}
	lang, err := util.GetString(loginInfo, submission.LANG)
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
