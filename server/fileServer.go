//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package server

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net"
)

type (
	//SubmissionSpawner is an implementation of
	//HandlerSpawner for SubmissionHandlers.
	SubmissionSpawner struct{}

	//SubmissionHandler is an implementation of ConnHandler
	//used to receive submissions from users of the impendulo system.
	SubmissionHandler struct {
		conn       net.Conn
		submission *project.Submission
	}
)

const (
	OK                  = "ok"
	SEND                = "send"
	LOGIN               = "begin"
	LOGOUT              = "end"
	REQ                 = "req"
	PROJECTS            = "projects"
	SUBMISSION_NEW      = "submission_new"
	SUBMISSION_CONTINUE = "submission_continue"
	LOG_SERVER          = "server/server.go"
)

//Spawn creates a new ConnHandler of type SubmissionHandler.
func (this *SubmissionSpawner) Spawn() ConnHandler {
	return &SubmissionHandler{}
}

//Start sets the connection, launches the Handle method
//and ends the session when it returns.
func (this *SubmissionHandler) Start(conn net.Conn) {
	this.conn = conn
	this.submission = new(project.Submission)
	this.submission.Id = bson.NewObjectId()
	this.End(this.Handle())
}

//End ends a session and reports any errors to the user.
func (this *SubmissionHandler) End(err error) {
	defer this.conn.Close()
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		util.Log(err)
	} else {
		msg = OK
	}
	this.write(msg)
}

//Handle manages a connection by authenticating it,
//processing its Submission and reading Files from it.
func (this *SubmissionHandler) Handle() (err error) {
	err = this.Login()
	if err != nil {
		return
	}
	err = this.LoadInfo()
	if err != nil {
		return
	}
	err = processing.StartSubmission(this.submission.Id)
	if err != nil {
		return
	}
	defer func() { processing.EndSubmission(this.submission.Id) }()
	done := false
	for err == nil && !done {
		done, err = this.Read()
	}
	return
}

//Login authenticates a Submission.
//It validates the user's credentials and permissions.
func (this *SubmissionHandler) Login() (err error) {
	loginInfo, err := util.ReadJson(this.conn)
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
	//Read user details
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
	u, err := db.User(this.submission.User)
	if err != nil {
		return
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("%q used an invalid username or password",
			this.submission.User)
		return
	}
	//Send a list of available projects to the user.
	projects, err := db.Projects(nil, bson.M{project.SKELETON: 0}, project.NAME)
	if err == nil {
		err = this.writeJson(projects)
	}
	return
}

//LoadInfo reads the Json request info.
//A new submission is then created or an existing one resumed
//depending on the request.
func (this *SubmissionHandler) LoadInfo() (err error) {
	reqInfo, err := util.ReadJson(this.conn)
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

//createSubmission is used when a client wishes to create a new submission.
//Submission info is read from the subInfo map and used to create a new
//submission in the db.
func (this *SubmissionHandler) createSubmission(subInfo map[string]interface{}) (err error) {
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
	this.submission.Status = project.BUSY
	err = db.Add(db.SUBMISSIONS, this.submission)
	if err == nil {
		err = this.writeJson(this.submission)
	}
	return
}

//continueSubmission is used when a client wishes to continue with a previous submission.
//The submission id is read from the subInfo map and then the submission os loaded from the db.
func (this *SubmissionHandler) continueSubmission(subInfo map[string]interface{}) (err error) {
	idStr, err := util.GetString(subInfo, project.SUBID)
	if err != nil {
		return
	}
	id, err := util.ReadId(idStr)
	if err != nil {
		return
	}
	change := bson.M{db.SET: bson.M{project.STATUS: project.BUSY}}
	err = db.Update(db.SUBMISSIONS, bson.M{project.ID: id}, change)
	if err != nil {
		return
	}
	this.submission, err = db.Submission(bson.M{project.ID: id}, nil)
	if err != nil {
		return
	}
	err = this.write(OK)
	return
}

//Read reads Files from the connection and sends them for processing.
func (this *SubmissionHandler) Read() (done bool, err error) {
	//Receive file metadata and request info
	requestInfo, err := util.ReadJson(this.conn)
	if err != nil {
		return
	}
	//Get the type of request
	req, err := util.GetString(requestInfo, REQ)
	if err != nil {
		return
	}
	if req == SEND {
		err = this.write(OK)
		if err != nil {
			return
		}
		//Receive file data
		var buffer []byte
		buffer, err = util.ReadData(this.conn)
		if err != nil {
			return
		}
		err = this.write(OK)
		if err != nil {
			return
		}
		delete(requestInfo, REQ)
		var file *project.File
		//Create a new file
		switch this.submission.Mode {
		case project.ARCHIVE_MODE:
			file = project.NewArchive(
				this.submission.Id, buffer)
		case project.FILE_MODE:
			file, err = project.NewFile(
				this.submission.Id, requestInfo, buffer)
			if err != nil {
				return
			}
		}
		err = db.Add(db.FILES, file)
		if err != nil {
			return
		}
		//Send file to be processed.
		err = processing.AddFile(file)
	} else if req == LOGOUT {
		//Logout request so we are done with this client.
		done = true
	} else {
		err = fmt.Errorf("Unknown request %q", req)
	}
	return
}

//writeJson writes an json data to this SubmissionHandler's connection.
func (this *SubmissionHandler) writeJson(data interface{}) (err error) {
	err = util.WriteJson(this.conn, data)
	if err == nil {
		_, err = this.conn.Write([]byte(util.EOT))
	}
	return
}

//write writes a string to this SubmissionHandler's connection.
func (this *SubmissionHandler) write(data string) (err error) {
	_, err = this.conn.Write([]byte(data))
	if err == nil {
		_, err = this.conn.Write([]byte(util.EOT))
	}
	return
}
