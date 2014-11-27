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

package receiver

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
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

	ProjectInfo struct {
		Project     *project.P
		Assignments []*AssignmentInfo
	}
	AssignmentInfo struct {
		Assignment  *project.Assignment
		Submissions []*project.Submission
	}
)

const (
	OK           = "ok"
	SEND         = "send"
	LOGIN        = "login"
	LOGOUT       = "logout"
	REQUEST      = "request"
	REGISTER     = "register"
	PROJECTS     = "projects"
	NEW          = "submission_new"
	CONTINUE     = "submission_continue"
	LOG_RECEIVER = "receiver/receiver.go"
)

func NewProjectInfo(p *project.P) *ProjectInfo {
	return &ProjectInfo{Project: p, Assignments: make([]*AssignmentInfo, 0, 1)}
}

func (p *ProjectInfo) Add(a *project.Assignment, subs []*project.Submission) {
	p.Assignments = append(p.Assignments, &AssignmentInfo{Assignment: a, Submissions: subs})
}

//Spawn creates a new ConnHandler of type SubmissionHandler.
func (s *SubmissionSpawner) Spawn() ConnHandler {
	return &SubmissionHandler{}
}

//Start sets the connection, launches the Handle method
//and ends the session when it returns.
func (s *SubmissionHandler) Start(c net.Conn) {
	s.conn = c
	s.submission = new(project.Submission)
	s.submission.Id = bson.NewObjectId()
	s.End(s.Handle())
}

//End ends a session and reports any errors to the user.
func (s *SubmissionHandler) End(e error) {
	defer s.conn.Close()
	msg := OK
	if e != nil {
		msg = "ERROR: " + e.Error()
		util.Log(e, LOG_RECEIVER)
	}
	s.write(msg)
}

//Handle manages a connection by authenticating it,
//processing its Submission and reading Files from it.
func (s *SubmissionHandler) Handle() error {
	var e error
	if e = s.Setup(); e != nil {
		return e
	}
	if e = s.LoadInfo(); e != nil {
		return e
	}
	if e = mq.StartSubmission(s.submission.Id); e != nil {
		return e
	}
	defer func() { mq.EndSubmission(s.submission.Id) }()
	d := false
	for !d {
		d, e = s.Read()
		if e != nil {
			return e
		}
	}
	return nil
}

//Setup initialises a Submission.
//It either logs a user in or registers a new user.
func (s *SubmissionHandler) Setup() error {
	j, e := util.ReadJSON(s.conn)
	if e != nil {
		return e
	}
	if e = s.login(j); e != nil {
		return e
	}
	m, e := convert.GetString(j, project.MODE)
	if e != nil {
		return e
	}
	if e = s.submission.SetMode(m); e != nil {
		return e
	}
	t := util.CurMilis()
	//Send a list of available projects to the user.
	as, e := db.Assignments(bson.M{db.START: bson.M{db.LT: t}, db.END: bson.M{db.GT: t}}, nil, db.PROJECTID)
	if e != nil {
		return e
	}
	pi := make([]*ProjectInfo, 0, len(as))
	for _, a := range as {
		if len(pi) == 0 || pi[len(pi)-1].Project.Id != a.ProjectId {
			p, e := db.Project(bson.M{db.ID: a.ProjectId}, nil)
			if e != nil {
				util.Log(e)
				continue
			}
			pi = append(pi, NewProjectInfo(p))
		}
		ss, e := db.Submissions(bson.M{db.USER: s.submission.User, db.ASSIGNMENTID: a.Id}, nil)
		if e != nil {
			util.Log(e)
		}
		pi[len(pi)-1].Add(a, ss)
	}
	return s.writeJSON(pi)
}

func (s *SubmissionHandler) login(m map[string]interface{}) error {
	r, e := convert.GetString(m, REQUEST)
	if e != nil {
		return e
	}
	un, e := convert.GetString(m, db.USER)
	if e != nil {
		return e
	}
	p, e := convert.GetString(m, user.PWORD)
	if e != nil {
		return e
	}
	switch r {
	case LOGIN:
		u, e := db.User(un)
		if e != nil {
			return e
		}
		if !util.Validate(u.Password, u.Salt, p) {
			return fmt.Errorf("%q used invalid username or password", un)
		}
	case REGISTER:
		if e = db.Add(db.USERS, user.New(un, p)); e != nil {
			return fmt.Errorf("user %s already exists", un)
		}
	default:
		return fmt.Errorf("invalid start request %q expected %s or %s", r, LOGIN, REGISTER)
	}
	s.submission.User = un
	return nil
}

//LoadInfo reads the Json request info.
//A new submission is then created or an existing one resumed
//depending on the request.
func (s *SubmissionHandler) LoadInfo() error {
	i, e := util.ReadJSON(s.conn)
	if e != nil {
		return e
	}
	r, e := convert.GetString(i, REQUEST)
	if e != nil {
		return e
	}
	switch r {
	case NEW:
		return s.createSubmission(i)
	case CONTINUE:
		return s.continueSubmission(i)
	}
	return fmt.Errorf("invalid request %q", r)
}

//createSubmission is used when a client wishes to create a new submission.
//Submission info is read from the subInfo map and used to create a new
//submission in the db.
func (s *SubmissionHandler) createSubmission(subInfo map[string]interface{}) error {
	var e error
	s.submission.AssignmentId, e = convert.GetId(subInfo, db.ASSIGNMENTID)
	if e != nil {
		return e
	}
	s.submission.ProjectId, e = convert.GetId(subInfo, db.PROJECTID)
	if e != nil {
		return e
	}
	s.submission.Time, e = convert.GetInt64(subInfo, db.TIME)
	if e != nil {
		return e
	}
	if e = db.Add(db.SUBMISSIONS, s.submission); e != nil {
		return e
	}
	return s.write(OK)
}

//continueSubmission is used when a client wishes to continue with a previous submission.
//The submission id is read from the subInfo map and then the submission os loaded from the db.
func (s *SubmissionHandler) continueSubmission(subInfo map[string]interface{}) error {
	v, e := convert.GetString(subInfo, db.SUBID)
	if e != nil {
		return e
	}
	id, e := convert.Id(v)
	if e != nil {
		return e
	}
	s.submission, e = db.Submission(bson.M{db.ID: id}, nil)
	if e != nil {
		return e
	}
	return s.write(OK)
}

//Read reads Files from the connection and sends them for processing.
func (s *SubmissionHandler) Read() (bool, error) {
	//Receive file metadata and request info
	i, e := util.ReadJSON(s.conn)
	if e != nil {
		return false, e
	}
	//Get the type of request
	r, e := convert.GetString(i, REQUEST)
	if e != nil {
		return false, e
	}
	switch r {
	case SEND:
		if e = s.write(OK); e != nil {
			return false, e
		}
		//Receive file data
		b, e := util.ReadData(s.conn)
		if e != nil {
			return false, e
		}
		if e = s.write(OK); e != nil {
			return false, e
		}
		delete(i, REQUEST)
		var f *project.File
		//Create a new file
		switch s.submission.Mode {
		case project.ARCHIVE_MODE:
			f = project.NewArchive(s.submission.Id, b)
		case project.FILE_MODE:
			if f, e = project.NewFile(s.submission.Id, i, b); e != nil {
				return false, e
			}
		}
		if e = db.Add(db.FILES, f); e != nil {
			return false, e
		}
		//Send file to be processed.
		return false, mq.AddFile(f)
	case LOGOUT:
		//Logout request so we are done with this client.
		return true, nil
	}
	return false, fmt.Errorf("Unknown request %q", r)
}

//writeJSON writes an JSON data to this SubmissionHandler's connection.
func (s *SubmissionHandler) writeJSON(i interface{}) error {
	if e := util.WriteJSON(s.conn, i); e != nil {
		return e
	}
	_, e := s.conn.Write([]byte(util.EOT))
	return e
}

//write writes a string to this SubmissionHandler's connection.
func (s *SubmissionHandler) write(d string) error {
	if _, e := s.conn.Write([]byte(d)); e != nil {
		return e
	}
	_, e := s.conn.Write([]byte(util.EOT))
	return e
}
