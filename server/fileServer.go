package server

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"io"
	"labix.org/v2/mgo/bson"
	"net"
)

//SubmissionSpawner is an implementation of HandlerSpawner for SubmissionHandlers.
type SubmissionSpawner struct {}

//Spawn creates a new ConnHandler of type SubmissionHandler.
func (this *SubmissionSpawner) Spawn() ConnHandler {
	return &SubmissionHandler{}
}

//SubmissionHandler is an implementation of ConnHandler used to receive submissions from users of the impendulo system.
type SubmissionHandler struct {
	Conn       net.Conn
	Submission *project.Submission
	fileCount int
}

//Start sets the connection, launches the Handle method and ends the session when it returns.
func (this *SubmissionHandler) Start(conn net.Conn) {
	this.Conn = conn
	this.Submission = new(project.Submission)
	this.Submission.Id = bson.NewObjectId()
	this.End(this.Handle())
}

//Handle manages a connection by authenticating it, processing its Submission and reading Files from it.
func (this *SubmissionHandler) Handle() error {
	err := this.Login()
	if err != nil {
		return err
	}
	err = this.LoadInfo()
	if err != nil {
		return err
	}
	processing.StartSubmission(this.Submission.Id)
	defer func() { processing.EndSubmission(this.Submission.Id) }()
	for err == nil {
		err = this.Read()
	}
	if err == io.EOF {
		return nil
	}
	return err
}

//Login authenticates this Submission by validating the user's credentials and permissions.
//This Submission is then stored.
func (this *SubmissionHandler) Login() error {
	loginInfo, err := util.ReadJson(this.Conn)
	if err != nil {
		return err
	}
	req, err := util.GetString(loginInfo, REQ)
	if err != nil {
		return err
	} else if req != LOGIN {
		return fmt.Errorf("Invalid request %q, expected %q", req, LOGIN)
	}
	this.Submission.User, err = util.GetString(loginInfo, project.USER)
	if err != nil {
		return err
	}
	pword, err := util.GetString(loginInfo, user.PWORD)
	if err != nil {
		return err
	}
	this.Submission.Mode, err = util.GetString(loginInfo, project.MODE)
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
	projects, err := db.GetProjects(nil)
	if err != nil {
		return err
	}
	return util.WriteJson(this.Conn, projects)
}

func (this *SubmissionHandler) LoadInfo() error {
	reqInfo, err := util.ReadJson(this.Conn)
	if err != nil {
		return err
	}
	req, err := util.GetString(reqInfo, REQ)
	if err != nil {
		return err
	} else if req == SUBMISSION_NEW {
		return this.createSubmission(reqInfo)
	} else if req == SUBMISSION_CONTINUE {
		return this.continueSubmission(reqInfo)
	} else {
		return fmt.Errorf("Invalid request %q", req)
	}
}

func (this *SubmissionHandler) createSubmission(subInfo map[string]interface{}) error {
	idStr, err := util.GetString(subInfo, project.PROJECT_ID)
	if err != nil {
		return err
	} else if !bson.IsObjectIdHex(idStr) {
		return fmt.Errorf("Invalid id hex %q", idStr)
	}
	this.Submission.ProjectId = bson.ObjectIdHex(idStr)
	this.Submission.Time, err = util.GetInt64(subInfo, project.TIME)
	if err != nil {
		return err
	}
	err = db.AddSubmission(this.Submission)
	if err != nil {
		return err
	}
	this.fileCount = 0
	return util.WriteJson(this.Conn, this.Submission)
}

func (this *SubmissionHandler) continueSubmission(subInfo map[string]interface{}) error {
	idStr, err := util.GetString(subInfo, project.SUBID)
	if err != nil {
		return err
	} else if !bson.IsObjectIdHex(idStr) {
		return fmt.Errorf("Invalid id hex %q", idStr)
	}
	this.Submission, err = db.GetSubmission(bson.M{project.ID: bson.ObjectIdHex(idStr)}, nil)
	if err != nil {
		return err
	}
	this.fileCount, err = db.Count(db.FILES, bson.M{project.SUBID: this.Submission.Id})
	if err != nil {
		return err
	}
	return util.WriteJson(this.Conn, this.fileCount)
}

//Read reads Files from the connection and sends them for processing.
func (this *SubmissionHandler) Read() (err error) {
	requestInfo, err := util.ReadJson(this.Conn)
	if err != nil {
		return
	}
	req, err := util.GetString(requestInfo, REQ)
	if err != nil {
		return
	}
	if req == SEND {
		_, err = this.Conn.Write([]byte(OK))
		if err != nil {
			return
		}
		var buffer []byte
		buffer, err = util.ReadData(this.Conn)
		if err != nil {
			return
		}
		_, err = this.Conn.Write([]byte(OK))
		if err != nil {
			return
		}
		delete(requestInfo, REQ)
		requestInfo[project.NUM] = this.fileCount
		this.fileCount ++
		var file *project.File
		file, err = project.NewFile(this.Submission.Id, requestInfo, buffer)
		if err != nil {
			return
		}
		err = db.AddFile(file)
		if err != nil {
			return
		}
		processing.AddFile(file)
	} else if req == LOGOUT {
		err = io.EOF
	} else{
		err = fmt.Errorf("Unknown request %q", req)
	}
	return
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
