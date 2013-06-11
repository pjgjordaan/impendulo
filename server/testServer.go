package server

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/project"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
	"net"
	"labix.org/v2/mgo/bson"
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
	Test *project.Test
}

//Start sets the connection, launches the Handle method and ends the session when it returns.
func (this *TestHandler) Start(conn net.Conn){
	this.Conn = conn
	this.Test = new(project.Test)
	this.End(this.Handle())
}

//Handle manages a connection by authenticating it and storing its submitted Test.
func (this *TestHandler) Handle() error {
	err := this.Login()
	if err != nil {
		return err
	}
	err = this.LoadInfo()
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
	this.Test.User, err = util.GetString(loginInfo, project.USER)
	if err != nil {
		return err
	}
	pword, err := util.GetString(loginInfo, user.PWORD)
	if err != nil {
		return err
	}
	u, err := db.GetUserById(this.Test.User)
	if err != nil {
		return err
	}
	if !u.CheckSubmit(project.TEST_MODE) {
		return fmt.Errorf("User %q has insufficient permissions for %q", this.Test.User, project.TEST_MODE)
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		return fmt.Errorf("User %q attempted to login with an invalid username or password", this.Test.User)
	}
	_, err = this.Conn.Write([]byte(OK))
	return err
}


func (this *TestHandler) LoadInfo() error{
	projects, err := db.GetProjects(bson.M{}, nil)
	if err != nil {
		return err
	}
	err = util.WriteJson(this.Conn, map[string]interface{}{"projects":projects})
	if err != nil {
		return err
	}
	info, err := util.ReadJSON(this.Conn)
	if err != nil {
		return err
	}
	this.Test.ProjectId, err = util.GetID(info, project.PROJECT_ID)
	if err != nil {
		return err
	}
	this.Test.Package, err = util.GetString(info, project.PKG)
	if err != nil {
		return err
	}
	this.Test.Name, err = util.GetString(info, project.NAME)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	return err
}

//Read retrieves information about the tests as well as the tests themselves and their data files from the connection.
func (this *TestHandler) Read() error{
	var err error
	this.Test.Test, err = util.ReadData(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
	this.Test.Data, err = util.ReadData(this.Conn)
	if err != nil {
		return err
	}
	_, err = this.Conn.Write([]byte(OK))
	if err != nil {
		return err
	}
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
