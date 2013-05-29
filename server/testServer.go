package server

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
	"net"
)

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
	pkg, err := util.GetString(testInfo, submission.PKG)
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
	this.Test = submission.NewTest(project, pkg, lang, names, testFiles, dataFiles) 
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
