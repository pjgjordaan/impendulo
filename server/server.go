package server

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
	"net"
	"io"
)

const (
	OK     = "ok"
	SEND   = "send"
	LOGIN  = "begin"
	LOGOUT = "end"
	REQ    = "req"
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
type HandlerSpawner interface{
	Spawn() ConnHandler
}

//ConnHandler is an interface with basic methods for handling connections.
type ConnHandler interface{
	Start(conn net.Conn)
	Handle() error
	Login() error
	Read() error
	End(err error)
}

//TestSpawner is a basic implementation of HandlerSpawner for TestHandlers.
type TestSpawner struct{}

//Spawn creates a new ConnHandler of type TestHandler.
func (this *TestSpawner) Spawn() ConnHandler{
	return new(TestHandler)
}

//TestHandler is an implementation of ConnHandler used to receive tests for projects.
type TestHandler struct{
	Conn net.Conn
	Test *submission.Test
}

//Start sets the connection, launches the Handle method and ends the session when it returns.
func (this *TestHandler) Start(conn net.Conn){
	this.Conn = conn
	this.End(this.Handle())
}

//Handle manages a connection by authenticating it and storing its submitted Test.
func (this *TestHandler) Handle() error {
	err := this.Login()
	if err != nil {
		return err
	}
	err = this.Read()
	if err != nil {
		return err
	}
	return db.AddTest(this.Test)
}

//Login authenticates a connection by validating user credentials and checking user permissions. 
func (this *TestHandler) Login() error {
	loginInfo, err := util.ReadJSON(this.Conn)
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
	_, err = this.Conn.Write([]byte(OK))
	return err
}

//Read retrieves information about the tests as well as the tests themselves and their data files from the connection.
func (this *TestHandler) Read() error{
	testInfo, err := util.ReadJSON(this.Conn)
	if err != nil {
		return err
	}
	project, err := util.GetString(testInfo, submission.PROJECT)
	if err != nil {
		return err
	}
	lang, err := util.GetString(testInfo, submission.LANG)
	if err != nil {
		return err
	}
	names, err := util.GetStrings(testInfo, submission.NAMES)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	testFiles, err := util.ReadData(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	dataFiles, err := util.ReadData(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	this.Test = submission.NewTest(project, lang, names, testFiles, dataFiles) 
	return nil
}

//End ends a session and reports any errors to the user. 
func (this *TestHandler) End(err error) {
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

//SubmissionSpawner is an implementation of HandlerSpawner for SubmissionHandlers.
type SubmissionSpawner struct{
	SubChan chan *submission.Submission
	FileChan chan *submission.File
}

//Spawn creates a new ConnHandler of type SubmissionHandler.
func (this *SubmissionSpawner) Spawn() ConnHandler{
	return &SubmissionHandler{SubChan: this.SubChan, FileChan: this.FileChan}
}

//SubmissionHandler is an implementation of ConnHandler used to receive submissions from users of the impendulo system.
type SubmissionHandler struct{
	Conn net.Conn
	Submission *submission.Submission
	//Used to send this Submission for processing.
	SubChan chan *submission.Submission
	//Used to send Files in this Submission for processing.
	FileChan chan *submission.File
}

//Start sets the connection, launches the Handle method and ends the session when it returns.
func (this *SubmissionHandler) Start(conn net.Conn){
	this.Conn = conn
	this.End(this.Handle())
}

//Handle manages a connection by authenticating it, processing its Submission and reading Files from it.
func (this *SubmissionHandler) Handle() error {
	err := this.Login()
	if err != nil {
		return err
	}
	this.SubChan <- this.Submission
	defer func(){this.SubChan <- this.Submission}()
	for err == nil{
		err = this.Read()
	} 
	if err == io.EOF{
		return nil
	}
	return err
}

//Login authenticates this Submission by validating the user's credentials and permissions.
//This Submission is then stored.
func (this *SubmissionHandler) Login() error {
	loginInfo, err := util.ReadJSON(this.Conn)
	if err != nil {
		return err
	}
	err = this.ReadSubmission(loginInfo)
	if err != nil {
		return err
	}
	pword, err := util.GetString(loginInfo, user.PWORD)
	if err != nil {
		return err
	}
	u, err := db.GetUserById(this.Submission.User)
	if err != nil {
		return err
	}
	if !u.CheckSubmit(this.Submission.Mode) {
		return fmt.Errorf("User %q has insufficient permissions for %q", this.Submission.User, this.Submission.Mode)
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		return fmt.Errorf("User %q attempted to login with an invalid username or password", this.Submission.User)
	}
	err = db.AddSubmission(this.Submission)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	return nil
}


//Read reads Files from the connection and sends them for processing.
func (this *SubmissionHandler) Read()error{
	requestInfo, err := util.ReadJSON(this.Conn)
	if err != nil {
		return err
	}
	req, err := util.GetString(requestInfo, REQ)
	if err != nil {
		return err
	}
	if req == SEND {
		_, err = this.Conn.Write([]byte(OK))
		if err != nil {
			return err
		}
		buffer, err := util.ReadData(this.Conn)
		if err != nil {
			return err
		}
		_, err = this.Conn.Write([]byte(OK))
		if err != nil {
			return err
		}
		delete(requestInfo, REQ)
		f := submission.NewFile(this.Submission.Id, requestInfo, buffer)
		err = db.AddFile(f)
		if err != nil {
			return err
		}
		this.FileChan <- f
		return nil
	} else if req == LOGOUT {
		return io.EOF
	} 
	return fmt.Errorf("Unknown request %q", req)
}

//End ends a session and reports any errors to the user. 
func (this *SubmissionHandler) End(err error) {
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

//ReadSubmission reads Submission values from a json object and creates a new Submission from them. 
func (this *SubmissionHandler) ReadSubmission(loginInfo map[string]interface{}) error {
	username, err := util.GetString(loginInfo, submission.USER)
	if err != nil {
		return err
	}
	project, err := util.GetString(loginInfo, submission.PROJECT)
	if err != nil {
		return err
	}
	mode, err := util.GetString(loginInfo, submission.MODE)
	if err != nil {
		return err
	}
	lang, err := util.GetString(loginInfo, submission.LANG)
	if err != nil {
		return err
	}
	this.Submission = submission.NewSubmission(project, username, mode, lang)
	return nil
}
