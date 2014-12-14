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

package context

import (
	"code.google.com/p/gorilla/sessions"

	"encoding/gob"
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/code"
	"github.com/godfried/impendulo/tool/result/description"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	//Context is used to keep track of the current user's session.
	C struct {
		Session *sessions.Session
		Browse  *Browse
	}

	//Browse is used to keep track of the user's browsing.
	Browse struct {
		IsUser          bool
		Pid, Aid, Sid   bson.ObjectId
		Uid, File, View string
		Current, Next   int
		DisplayCount    int
		Level           Level
		Result          *description.D
	}
	Level  int
	Setter func(*http.Request) error
)

const (
	STORE_NAME       = "impendulo"
	_          Level = iota
	HOME
	PROJECTS
	USERS
	ASSIGNMENTS
	SUBMISSIONS
	FILES
	ANALYSIS
)

var (
	store sessions.Store
)

func init() {
	if e := loadStore(); e != nil {
		panic(e)
	}
	//Register these so that they can be saved with the session.
	gob.Register(new(Browse))
	gob.Register(new(bson.ObjectId))
}

func loadStore() error {
	if store != nil {
		return nil
	}
	auth, enc, e := util.CookieKeys()
	if e != nil {
		return e
	}
	store = sessions.NewCookieStore(auth, enc)
	return nil
}

func loadSession(r *http.Request, n string) (*sessions.Session, error) {
	if e := loadStore(); e != nil {
		return nil, e
	}
	return store.Get(r, n)
}

func Load(r *http.Request) (*C, error) {
	return LoadN(r, STORE_NAME)
}

func LoadN(r *http.Request, n string) (*C, error) {
	s, e := loadSession(r, n)
	if e != nil {
		return nil, e
	}
	return load(s), nil
}

//Close closes a session.
func (c *C) Close() {
	c.save()
}

//save
func (c *C) save() {
	c.Session.Values["browse"] = c.Browse
}

//Save stores the current session.
func (c *C) Save(r *http.Request, w http.ResponseWriter) error {
	c.save()
	return c.Session.Save(r, w)
}

//IsView checks whether the given view matches the user's current view.
func (c *C) IsView(v string) bool {
	return c.Browse.View == v
}

//LoggedIn checks whether a user is signed in.
func (c *C) LoggedIn() bool {
	_, e := c.Username()
	return e == nil
}

//Username retrieves the current user's username.
func (c *C) Username() (string, error) {
	u, ok := c.Session.Values["user"].(string)
	if !ok {
		return "", fmt.Errorf("could not retrieve user")
	}
	return u, nil
}

//AddUser sets the currently signed in user.
func (c *C) AddUser(u string) {
	c.Session.Values["user"] = u
}

//AddUser sets the currently signed in user.
func (c *C) RemoveUser() {
	delete(c.Session.Values, "user")
}

//AddMessage adds a message to be displayed to the user.
func (c *C) AddMessage(m string, isErr bool) {
	if isErr {
		c.Session.AddFlash(m, "error")
	} else {
		c.Session.AddFlash(m, "success")
	}
}

//Errors retrieves all error messages.
func (c *C) Errors() []interface{} {
	return c.Session.Flashes("error")
}

//Successes retrieves all success messages.
func (c *C) Successes() []interface{} {
	return c.Session.Flashes("success")
}

//Load loads a context from the session.
func load(s *sessions.Session) *C {
	c := &C{Session: s}
	if v, ok := c.Session.Values["browse"]; ok {
		c.Browse = v.(*Browse)
	} else {
		c.Browse = new(Browse)
		c.Browse.DisplayCount = 10
		c.Browse.Current = 0
		c.Browse.Next = 0
		c.Browse.Result = &description.D{Type: code.NAME}
	}
	u, e := c.Username()
	if e != nil {
		return c
	}
	if _, e = db.User(u); e != nil {
		c.RemoveUser()
	}
	return c
}

func (b *Browse) ClearProject() {
	b.Pid = ""
	b.ClearAssignment()
}

func (b *Browse) ClearUser() {
	b.Uid = ""
	b.ClearAssignment()
}

func (b *Browse) ClearAssignment() {
	b.Aid = ""
	b.ClearSubmission()
}

func (b *Browse) ClearSubmission() {
	b.File = ""
	b.Current = 0
	b.Next = 0
	b.DisplayCount = 10
	if b.Result == nil {
		b.Result = &description.D{Type: code.NAME}
	}
	b.Sid = ""
}

func (b *Browse) SetDisplayCount(r *http.Request) error {
	i, e := webutil.Int(r, "displaycount")
	if e == nil {
		b.DisplayCount = i + 10
	} else {
		b.DisplayCount = 10
	}
	return nil
}

func (b *Browse) SetUid(r *http.Request) error {
	id, e := webutil.String(r, "user-id")
	if e != nil {
		return nil
	}
	b.ClearUser()
	b.IsUser = true
	b.Uid = id
	return nil
}

func (b *Browse) SetPid(r *http.Request) error {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return nil
	}
	b.ClearProject()
	b.IsUser = false
	b.Pid = pid
	return nil
}

func (b *Browse) SetAid(r *http.Request) error {
	id, e := webutil.Id(r, "assignment-id")
	if e != nil {
		return nil
	}
	b.ClearAssignment()
	a, e := db.Assignment(bson.M{db.ID: id}, nil)
	if e != nil {
		return e
	}
	p, e := db.Project(bson.M{db.ID: a.ProjectId}, nil)
	if e != nil {
		return e
	}
	b.Aid = id
	b.Pid = p.Id
	b.File = p.Name + ".java"
	return nil
}

func (b *Browse) SetSid(r *http.Request) error {
	id, e := webutil.Id(r, "submission-id")
	if e != nil {
		return nil
	}
	b.ClearSubmission()
	s, e := db.Submission(bson.M{db.ID: id}, nil)
	if e != nil {
		return e
	}
	a, e := db.Assignment(bson.M{db.ID: s.AssignmentId}, nil)
	if e != nil {
		return e
	}
	p, e := db.Project(bson.M{db.ID: a.ProjectId}, nil)
	if e != nil {
		return e
	}
	b.Sid = id
	b.Aid = a.Id
	b.Pid = p.Id
	b.Uid = s.User
	b.File = p.Name + ".java"
	return b.Result.Update(b.Sid, b.File)
}

func (b *Browse) SetResult(r *http.Request) error {
	s, e := webutil.String(r, "result")
	if e != nil {
		return nil
	}
	return b.Result.Set(s)
}

func (b *Browse) SetFile(r *http.Request) error {
	n, e := webutil.String(r, "file")
	if e != nil {
		return nil
	}
	b.File = n
	b.Current = 0
	b.Next = 0
	return nil
}

func (b *Browse) SetIsUser(r *http.Request) error {
	u, e := webutil.Bool(r, "isuser")
	if e != nil {
		return nil
	}
	b.IsUser = u
	return nil
}

func (b *Browse) SetFileIndices(r *http.Request) error {
	if b.File == "" {
		return nil
	}
	fs, e := db.Snapshots(b.Sid, b.File)
	if e != nil {
		util.Log(e)
		return nil
	}
	if e = b.setIndices(r, fs); e != nil {
		return b.setTimeIndices(r, fs)
	}
	return nil
}

func (b *Browse) setIndices(r *http.Request, fs []*project.File) error {
	defer func() {
		if b.Current == b.Next {
			if b.Current == 0 {
				b.Next = len(fs) - 1
			} else {
				b.Next = (b.Current + 1) % len(fs)
			}
		}
	}()
	c, e := webutil.Index(r, "current", len(fs)-1)
	if e != nil {
		return e
	}
	n, e := webutil.Index(r, "next", len(fs)-1)
	if e != nil {
		return e
	}
	b.Current = c
	b.Next = n
	return nil
}

func (b *Browse) setTimeIndices(r *http.Request, fs []*project.File) error {
	t, e := webutil.Int64(r, "time")
	if e != nil {
		return nil
	}
	for i, f := range fs {
		if f.Time == t {
			b.Current = i
			b.Next = (i + 1) % len(fs)
			return nil
		}
	}
	return fmt.Errorf("no file found at time %d", t)
}

func (b *Browse) Update(r *http.Request) error {
	setters := []Setter{b.SetIsUser, b.SetPid, b.SetUid, b.SetAid, b.SetSid, b.SetFile, b.SetResult, b.SetFileIndices, b.SetDisplayCount}
	for _, s := range setters {
		if e := s(r); e != nil {
			return e
		}
	}
	return nil
}

func (b *Browse) Submissions() ([]*project.Submission, error) {
	m := bson.M{}
	if b.IsUser {
		m[db.USER] = b.Uid
	} else {
		m[db.PROJECTID] = b.Pid
	}
	return db.Submissions(m, nil, "-"+db.TIME)
}

func (b *Browse) SetLevel(route string) {
	switch route {
	case "homeview":
		b.Level = HOME
	case "overviewresult", "overviewchart":
		if b.IsUser {
			b.Level = USERS
		} else {
			b.Level = PROJECTS
		}
	case "assignmentsview", "assignmentschart":
		b.Level = ASSIGNMENTS
	case "submissionsview", "submissionschart":
		b.Level = SUBMISSIONS
	case "filesview":
		b.Level = FILES
	case "resultsview":
		b.Level = ANALYSIS
	default:
	}
}

func (l Level) String() string {
	switch l {
	case HOME:
		return "home"
	case PROJECTS:
		return "projects"
	case USERS:
		return "users"
	case SUBMISSIONS:
		return "submissions"
	case ASSIGNMENTS:
		return "assignments"
	case FILES:
		return "files"
	case ANALYSIS:
		return "analysis"
	default:
		return "unknown"
	}
}

func (l Level) Is(level string) bool {
	return l.String() == level
}
