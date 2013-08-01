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
)

//SubmissionSpawner is an implementation of HandlerSpawner for SubmissionHandlers.
type SubmissionSpawner struct{}

//Spawn creates a new ConnHandler of type SubmissionHandler.
func (this *SubmissionSpawner) Spawn() RWCHandler {
	return &SubmissionHandler{}
}

//SubmissionHandler is an implementation of ConnHandler used to receive submissions from users of the impendulo system.
type SubmissionHandler struct {
	rwc       io.ReadWriteCloser
	submission *project.Submission
	fileCount  int
}

//Start sets the connection, launches the Handle method and ends the session when it returns.
func (this *SubmissionHandler) Start(rwc io.ReadWriteCloser) {
	this.rwc = rwc
	this.submission = new(project.Submission)
	this.submission.Id = bson.NewObjectId()
	this.End(this.Handle())
}

//End ends a session and reports any errors to the user.
func (this *SubmissionHandler) End(err error) {
	defer this.rwc.Close()
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		util.Log(err)
	} else {
		msg = OK
	}
	this.rwc.Write([]byte(msg))
}

//Handle manages a connection by authenticating it, processing its Submission and reading Files from it.
func(this *SubmissionHandler) Handle() (err error) {
	err = this.Login()
	if err != nil {
		return
	} 
	err = this.LoadInfo()
	if err != nil {
		return
	}
	processing.StartSubmission(this.submission.Id)
	defer func() { processing.EndSubmission(this.submission.Id) }()
	done := false
	for err == nil && !done {
		done, err = this.Read()
	}
	return
}

//Login authenticates this Submission by validating the user's credentials and permissions.
//This Submission is then stored.
func (this *SubmissionHandler) Login() (err error) {
	loginInfo, err := util.ReadJson(this.rwc)
	if err != nil {
		return
	}
	req, err := util.GetString(loginInfo, REQ)
	if err != nil {
		return
	} else if req != LOGIN {
		err = fmt.Errorf("Invalid request %q, expected %q", req, LOGIN)
		return
	}
	this.submission.User, err = util.GetString(loginInfo, project.USER)
	if err != nil {
		return
	}
	pword, err := util.GetString(loginInfo, user.PWORD)
	if err != nil {
		return
	}
	mode, err := util.GetString(loginInfo, project.MODE)
	if err != nil {
		return
	}
	err = this.submission.SetMode(mode)
	if err != nil {
		return
	}
	u, err := db.GetUserById(this.submission.User)
	if err != nil {
		return
	}
	if !u.CheckSubmit(this.submission.Mode) {
		err = fmt.Errorf("User %q has insufficient permissions for %q", this.submission.User, this.submission.Mode)
		return
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("User %q attempted to login with an invalid username or password", this.submission.User)
		return
	} 
	projects, err := db.GetProjects(nil)
	if err != nil {
		return
	}
	err = util.WriteJson(this.rwc, projects)
	return
}

func (this *SubmissionHandler) LoadInfo() (err error) {
	reqInfo, err := util.ReadJson(this.rwc)
	if err != nil {
		return
	}
	req, err := util.GetString(reqInfo, REQ)
	if err != nil {
		return
	} else if req == SUBMISSION_NEW {
		err = this.createSubmission(reqInfo)
	} else if req == SUBMISSION_CONTINUE {
		err = this.continueSubmission(reqInfo)
	} else {
		err = fmt.Errorf("Invalid request %q", req)
	}
	return
}


func (this *SubmissionHandler) createSubmission(subInfo map[string]interface{})(err error) {
	idStr, err := util.GetString(subInfo, project.PROJECT_ID)
	if err != nil {
		return
	} 
	this.submission.ProjectId, err = util.ReadId(idStr)
	if err != nil {
		return
	}
	this.submission.Time, err = util.GetInt64(subInfo, project.TIME)
	if err != nil {
		return
	}
	err = db.AddSubmission(this.submission)
	if err != nil {
		return
	}
	this.fileCount = 0
	err = util.WriteJson(this.rwc, this.submission)
	return
}

func (this *SubmissionHandler) continueSubmission(subInfo map[string]interface{})(err error) {
	idStr, err := util.GetString(subInfo, project.SUBID)
	if err != nil {
		return
	}
	id, err := util.ReadId(idStr)
	if err != nil {
		return
	}
	this.submission, err = db.GetSubmission(bson.M{project.ID: id}, nil)
	if err != nil {
		return err
	}
	this.fileCount, err = db.Count(db.FILES, bson.M{project.SUBID: this.submission.Id})
	if err != nil {
		return err
	}
	return util.WriteJson(this.rwc, this.fileCount)
}

//Read reads Files from the connection and sends them for processing.
func (this *SubmissionHandler) Read() (done bool, err error) {
	requestInfo, err := util.ReadJson(this.rwc)
	if err != nil {
		return
	}
	req, err := util.GetString(requestInfo, REQ)
	if err != nil {
		return
	}
	if req == SEND {
		_, err = this.rwc.Write([]byte(OK))
		if err != nil {
			return
		}
		var buffer []byte
		buffer, err = util.ReadData(this.rwc)
		if err != nil {
			return
		}
		_, err = this.rwc.Write([]byte(OK))
		if err != nil {
			return
		}
		delete(requestInfo, REQ)
		requestInfo[project.NUM] = this.fileCount
		this.fileCount++
		var file *project.File
		switch this.submission.Mode{
		case project.ARCHIVE_MODE:
			file = project.NewArchive(this.submission.Id, buffer, project.ZIP)
		case project.FILE_MODE:
			file, err = project.NewFile(this.submission.Id, requestInfo, buffer)
		}
		if err != nil {
			return
		}
		err = db.AddFile(file)
		if err != nil {
			return
		}
		processing.AddFile(file)
	} else if req == LOGOUT {
		done = true
	} else {
		err = fmt.Errorf("Unknown request %q", req)
	}
	return
}
