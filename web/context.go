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

package web

import (
	"code.google.com/p/gorilla/sessions"
	"encoding/gob"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
)

type (
	//Context is used to keep track of the current user's session.
	Context struct {
		Session    *sessions.Session
		Browse     *Browse
		Additional *Browse
	}

	//Browse is used to keep track of the user's browsing.
	Browse struct {
		IsUser                     bool
		Pid, Sid                   bson.ObjectId
		Uid, File, Result, View    string
		ChildFile                  string
		Current, Next              int
		CurrentChild, DisplayCount int
		Level                      Level
		Type, ChildType            project.Type
	}
	Level  int
	Setter func(*http.Request) error
)

const (
	HOME Level = iota
	PROJECTS
	USERS
	SUBMISSIONS
	FILES
	ANALYSIS
	CHART
)

func init() {
	//Register these so that they can be saved with the session.
	gob.Register(new(Browse))
	gob.Register(new(bson.ObjectId))
}

//Close closes a session.
func (c *Context) Close() {
	c.save()
}

//save
func (c *Context) save() {
	c.Session.Values["browse"] = c.Browse
}

//Save stores the current session.
func (c *Context) Save(req *http.Request, buff *HttpBuffer) error {
	c.save()
	return c.Session.Save(req, buff)
}

//IsView checks whether the given view matches the user's current view.
func (c *Context) IsView(v string) bool {
	return c.Browse.View == v
}

//LoggedIn checks whether a user is signed in.
func (c *Context) LoggedIn() bool {
	_, err := c.Username()
	return err == nil
}

//Username retrieves the current user's username.
func (c *Context) Username() (string, error) {
	u, ok := c.Session.Values["user"].(string)
	if !ok {
		return "", fmt.Errorf("Could not retrieve user.")
	}
	return u, nil
}

//AddUser sets the currently signed in user.
func (c *Context) AddUser(u string) {
	c.Session.Values["user"] = u
}

//AddUser sets the currently signed in user.
func (c *Context) RemoveUser() {
	delete(c.Session.Values, "user")
}

//AddMessage adds a message to be displayed to the user.
func (c *Context) AddMessage(m string, isErr bool) {
	var t string
	if isErr {
		t = "error"
	} else {
		t = "success"
	}
	c.Session.AddFlash(m, t)
}

//Errors retrieves all error messages.
func (c *Context) Errors() []interface{} {
	return c.Session.Flashes("error")
}

//Successes retrieves all success messages.
func (c *Context) Successes() []interface{} {
	return c.Session.Flashes("success")
}

//LoadContext loads a context from the session.
func LoadContext(sess *sessions.Session) *Context {
	c := &Context{Session: sess}
	if v, ok := c.Session.Values["browse"]; ok {
		c.Browse = v.(*Browse)
	} else {
		c.Browse = new(Browse)
		c.Browse.DisplayCount = 10
		c.Browse.Current = 0
		c.Browse.Next = 0
		c.Browse.CurrentChild = 0
	}
	u, err := c.Username()
	if err != nil {
		return c
	}
	_, err = db.User(u)
	if err != nil {
		c.RemoveUser()
	}
	return c
}

func (b *Browse) ClearSubmission() {
	b.File = ""
	b.ChildFile = ""
	b.Result = ""
	b.Current = 0
	b.Next = 0
	b.CurrentChild = 0
	b.DisplayCount = 10
	b.Type = project.SRC
	b.ChildType = project.SRC
}

func (b *Browse) SetDisplayCount(req *http.Request) error {
	i, err := GetInt(req, "displaycount")
	if err == nil {
		b.DisplayCount = i + 10
	} else {
		b.DisplayCount = 10
	}
	return nil
}

func (b *Browse) SetUid(req *http.Request) error {
	id, _, err := getUserId(req)
	if err != nil {
		return nil
	}
	b.IsUser = true
	b.Uid = id
	return nil
}

func (b *Browse) SetSid(req *http.Request) error {
	id, _, err := getSubId(req)
	if err != nil {
		return nil
	}
	sub, err := db.Submission(bson.M{db.ID: id}, bson.M{db.PROJECTID: 1, db.USER: 1})
	if err != nil {
		return err
	}
	b.ClearSubmission()
	b.Sid = id
	b.Pid = sub.ProjectId
	b.Uid = sub.User
	return nil
}

func (b *Browse) SetPid(req *http.Request) error {
	pid, _, err := getProjectId(req)
	if err != nil {
		return nil
	}
	b.IsUser = false
	b.Pid = pid
	return nil
}

func (b *Browse) SetResult(req *http.Request) error {
	r, err := GetString(req, "result")
	if err != nil {
		return nil
	}
	b.Result = r
	return nil
}

func (b *Browse) SetFile(req *http.Request) error {
	n, err := GetString(req, "file")
	if err != nil {
		return nil
	}
	b.File = n
	f, err := db.File(bson.M{db.SUBID: b.Sid, db.NAME: n}, bson.M{db.TYPE: 1})
	if err != nil {
		return err
	}
	b.Type = f.Type
	b.Current = 0
	b.Next = 0
	return nil
}

func (b *Browse) setIndices(req *http.Request, files []*project.File) error {
	defer func() {
		if b.Current == b.Next {
			b.Next = (b.Current + 1) % len(files)
		}
	}()
	c, err := getCurrent(req, len(files)-1)
	if err != nil {
		return err
	}
	n, err := getNext(req, len(files)-1)
	if err != nil {
		return err
	}
	b.Current = c
	b.Next = n
	return nil
}

func (b *Browse) setTimeIndices(req *http.Request, files []*project.File) error {
	t, err := strconv.ParseInt(req.FormValue("time"), 10, 64)
	if err != nil {
		return nil
	}
	for i, f := range files {
		if f.Time == t {
			b.Current = i
			b.Next = (i + 1) % len(files)
			return nil
		}
	}
	return fmt.Errorf("no file found at time %d", t)
}

func (b *Browse) SetFileIndices(req *http.Request) error {
	if b.File == "" {
		return nil
	}
	files, err := Snapshots(b.Sid, b.File, b.Type)
	if err != nil {
		return err
	}

	if err = b.setIndices(req, files); err != nil {
		return b.setTimeIndices(req, files)
	}
	return nil
}

func (b *Browse) SetChild(req *http.Request) error {
	if b.ChildFile == "" {
		f, err := testedFileName(b.Sid)
		if err != nil {
			return nil
		}
		b.ChildFile = f
		b.ChildType = project.SRC
		b.CurrentChild = 0
	}
	files, err := Snapshots(b.Sid, b.ChildFile, b.ChildType)
	if err != nil {
		return err
	}
	if i, err := getIndex(req, "currentchild", len(files)-1); err == nil {
		b.CurrentChild = i
	}
	return nil
}

func (b *Browse) Update(req *http.Request) error {
	setters := []Setter{b.SetPid, b.SetUid, b.SetSid, b.SetResult, b.SetFile, b.SetFileIndices, b.SetDisplayCount, b.SetChild}
	for _, s := range setters {
		if err := s(req); err != nil {
			return err
		}
	}
	return nil
}

func (b *Browse) childResult() bool {
	return b.Type == project.TEST && (b.Result == junit.NAME || b.Result == jacoco.NAME)
}

func (b *Browse) SetLevel(route string) {
	switch route {
	case "homeview":
		b.Level = HOME
	case "projectresult":
		b.Level = PROJECTS
	case "userresult":
		b.Level = USERS
	case "getsubmissions":
		b.Level = SUBMISSIONS
	case "getfiles":
		b.Level = FILES
	case "getsubmissionschart":
		b.Level = SUBMISSIONS
	case "displaychart":
		b.Level = CHART
	case "displayresult":
		b.Level = ANALYSIS

	}
}

func (this Level) Is(level string) bool {
	level = strings.ToLower(level)
	switch level {
	case "home":
		return this == HOME
	case "projects":
		return this == PROJECTS
	case "users":
		return this == USERS
	case "submissions":
		return this == SUBMISSIONS
	case "files":
		return this == FILES
	case "analysis":
		return this == ANALYSIS
	case "chart":
		return this == CHART
	default:
		return false
	}
}
