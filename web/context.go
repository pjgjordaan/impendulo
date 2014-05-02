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
	"errors"
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
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
		IsUser          bool
		Pid, Sid        bson.ObjectId
		Uid, File, View string
		//ChildFile                  string
		Current, Next int
		DisplayCount  int
		Level         Level
		Type          project.Type
		Result        *ResultDesc
	}
	ResultDesc struct {
		Type   string
		Name   string
		FileID bson.ObjectId
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
func (c *Context) Save(r *http.Request, b *HttpBuffer) error {
	c.save()
	return c.Session.Save(r, b)
}

//IsView checks whether the given view matches the user's current view.
func (c *Context) IsView(v string) bool {
	return c.Browse.View == v
}

//LoggedIn checks whether a user is signed in.
func (c *Context) LoggedIn() bool {
	_, e := c.Username()
	return e == nil
}

//Username retrieves the current user's username.
func (c *Context) Username() (string, error) {
	u, ok := c.Session.Values["user"].(string)
	if !ok {
		return "", fmt.Errorf("could not retrieve user")
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
	if isErr {
		c.Session.AddFlash(m, "error")
	} else {
		c.Session.AddFlash(m, "success")
	}
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
func LoadContext(s *sessions.Session) *Context {
	c := &Context{Session: s}
	if v, ok := c.Session.Values["browse"]; ok {
		c.Browse = v.(*Browse)
	} else {
		c.Browse = new(Browse)
		c.Browse.DisplayCount = 10
		c.Browse.Current = 0
		c.Browse.Next = 0
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

func (b *Browse) Src() (string, error) {
	if b.Type == project.SRC && b.File != "" {
		return b.File, nil
	}
	return "", errors.New("no source file available")
}

func (b *Browse) Test() (string, error) {
	if b.Type == project.TEST && b.File != "" {
		return b.File, nil
	}
	return "", errors.New("no source file available")
}

func (b *Browse) ClearSubmission() {
	b.File = ""
	b.Result = &ResultDesc{}
	b.Current = 0
	b.Next = 0
	b.DisplayCount = 10
	b.Type = project.SRC
}

func (b *Browse) SetDisplayCount(r *http.Request) error {
	i, e := convert.Int(r.FormValue("displaycount"))
	if e == nil {
		b.DisplayCount = i + 10
	} else {
		b.DisplayCount = 10
	}
	return nil
}

func (b *Browse) SetUid(r *http.Request) error {
	id, e := GetString(r, "user-id")
	if e != nil {
		return nil
	}
	b.IsUser = true
	b.Uid = id
	return nil
}

func (b *Browse) SetSid(r *http.Request) error {
	id, e := convert.Id(r.FormValue("submission-id"))
	if e != nil {
		return nil
	}
	s, e := db.Submission(bson.M{db.ID: id}, bson.M{db.PROJECTID: 1, db.USER: 1})
	if e != nil {
		return e
	}
	b.ClearSubmission()
	b.Sid = id
	b.Pid = s.ProjectId
	b.Uid = s.User
	return nil
}

func (b *Browse) SetPid(r *http.Request) error {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return nil
	}
	b.IsUser = false
	b.Pid = pid
	return nil
}

func (b *Browse) SetResult(r *http.Request) error {
	s, e := GetString(r, "result")
	if e != nil {
		return nil
	}
	return b.Result.Set(s)
}

func (b *Browse) SetFile(r *http.Request) error {
	n, e := GetString(r, "file")
	if e != nil {
		return nil
	}
	f, e := db.File(bson.M{db.SUBID: b.Sid, db.NAME: n}, bson.M{db.TYPE: 1})
	if e != nil {
		return e
	}
	b.File = n
	b.Type = f.Type
	b.Current = 0
	b.Next = 0
	return nil
}

func (b *Browse) setIndices(r *http.Request, fs []*project.File) error {
	defer func() {
		if b.Current == b.Next {
			b.Next = (b.Current + 1) % len(fs)
		}
	}()
	c, e := getIndex(r, "current", len(fs)-1)
	if e != nil {
		return e
	}
	n, e := getIndex(r, "next", len(fs)-1)
	if e != nil {
		return e
	}
	b.Current = c
	b.Next = n
	return nil
}

func (b *Browse) setTimeIndices(r *http.Request, fs []*project.File) error {
	t, e := strconv.ParseInt(r.FormValue("time"), 10, 64)
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

func (b *Browse) SetFileIndices(r *http.Request) error {
	if b.File == "" {
		return nil
	}
	fs, e := Snapshots(b.Sid, b.File, b.Type)
	if e != nil {
		return e
	}
	if e = b.setIndices(r, fs); e != nil {
		return b.setTimeIndices(r, fs)
	}
	return nil
}

func (b *Browse) Update(r *http.Request) error {
	setters := []Setter{b.SetPid, b.SetUid, b.SetSid, b.SetResult, b.SetFile, b.SetFileIndices, b.SetDisplayCount}
	for _, s := range setters {
		if e := s(r); e != nil {
			return e
		}
	}
	return nil
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

func (l Level) Is(level string) bool {
	level = strings.ToLower(level)
	switch level {
	case "home":
		return l == HOME
	case "projects":
		return l == PROJECTS
	case "users":
		return l == USERS
	case "submissions":
		return l == SUBMISSIONS
	case "files":
		return l == FILES
	case "analysis":
		return l == ANALYSIS
	case "chart":
		return l == CHART
	default:
		return false
	}
}

func NewResultDesc(s string) (*ResultDesc, error) {
	r := &ResultDesc{}
	if e := r.Set(s); e != nil {
		return nil, e
	}
	return r, nil
}

func (r *ResultDesc) Set(s string) error {
	r.Type = ""
	r.Name = ""
	r.FileID = ""
	sp := strings.Split(s, "-")
	if len(sp) > 1 {
		id, e := convert.Id(sp[1])
		if e != nil {
			return e
		}
		r.FileID = id
	}
	sp = strings.Split(sp[0], ":")
	if len(sp) > 1 {
		r.Name = sp[1]
	}
	r.Type = sp[0]
	return nil
}

func (r *ResultDesc) Format() string {
	s := r.Type
	if r.Name == "" {
		return s
	}
	s += " \u2192 " + r.Name
	if r.FileID == "" {
		return s
	}
	f, e := db.File(bson.M{db.ID: r.FileID}, bson.M{db.TIME: 1})
	if e != nil {
		return s + " \u2192 No File Found"
	}
	return s + " \u2192 " + util.Date(f.Time)
}

func (r *ResultDesc) Raw() string {
	s := r.Type
	if r.Name == "" {
		return r.Type
	}
	s += ":" + r.Name
	if r.FileID == "" {
		return s
	}
	return s + "-" + r.FileID.Hex()
}

func (r *ResultDesc) HasCode() bool {
	return r.Name != ""
}
