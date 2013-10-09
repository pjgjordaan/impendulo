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

package webserver

import (
	"code.google.com/p/gorilla/sessions"
	"encoding/gob"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
)

type (
	//Context is used to keep track of the current user's session.
	Context struct {
		Session *sessions.Session
		Browse  *Browse
	}

	//Browse is used to keep track of the user's browsing.
	Browse struct {
		IsUser                  bool
		Pid, Sid                bson.ObjectId
		Uid, File, Result, View string
		Current, Next           int
		Level                   Level
	}
	Level int
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
func (ctx *Context) Close() {
	ctx.save()
}

//save
func (ctx *Context) save() {
	ctx.Session.Values["browse"] = ctx.Browse
}

//Save stores the current session.
func (ctx *Context) Save(req *http.Request, buff *HttpBuffer) error {
	ctx.save()
	return ctx.Session.Save(req, buff)
}

//IsView checks whether the given view matches the user's current view.
func (ctx *Context) IsView(view string) bool {
	return ctx.Browse.View == view
}

//LoggedIn checks whether a user is signed in.
func (ctx *Context) LoggedIn() bool {
	_, err := ctx.Username()
	return err == nil
}

//Username retrieves the current user's username.
func (ctx *Context) Username() (string, error) {
	username, ok := ctx.Session.Values["user"].(string)
	if !ok {
		return "", fmt.Errorf("Could not retrieve user.")
	}
	return username, nil
}

//AddUser sets the currently signed in user.
func (ctx *Context) AddUser(user string) {
	ctx.Session.Values["user"] = user
}

//AddUser sets the currently signed in user.
func (ctx *Context) RemoveUser() {
	delete(ctx.Session.Values, "user")
}

//AddMessage adds a message to be displayed to the user.
func (ctx *Context) AddMessage(msg string, isErr bool) {
	var tipe string
	if isErr {
		tipe = "error"
	} else {
		tipe = "success"
	}
	ctx.Session.AddFlash(msg, tipe)
}

//Errors retrieves all error messages.
func (ctx *Context) Errors() []interface{} {
	return ctx.Session.Flashes("error")
}

//Successes retrieves all success messages.
func (ctx *Context) Successes() []interface{} {
	return ctx.Session.Flashes("success")
}

//LoadContext loads a context from the session.
func LoadContext(sess *sessions.Session) *Context {
	ctx := &Context{Session: sess}
	if val, ok := ctx.Session.Values["browse"]; ok {
		ctx.Browse = val.(*Browse)
	} else {
		ctx.Browse = new(Browse)
	}
	if uname, err := ctx.Username(); err == nil {
		_, err = db.User(uname)
		if err != nil {
			ctx.RemoveUser()
		}
	}
	return ctx
}

func (b *Browse) SetUid(req *http.Request) (err error) {
	uid, err := GetString(req, "uid")
	if err == nil {
		b.Uid = uid
	}
	return
}

func (b *Browse) SetSid(req *http.Request) (err error) {
	sid, err := util.ReadId(req.FormValue("sid"))
	if err == nil {
		b.Sid = sid
	}
	return
}

func (b *Browse) SetPid(req *http.Request) (err error) {
	pid, err := util.ReadId(req.FormValue("pid"))
	if err == nil {
		b.Pid = pid
	}
	return
}

func (b *Browse) SetResult(req *http.Request) (err error) {
	res, err := GetString(req, "result")
	if err == nil {
		b.Result = res
	}
	return
}

func (b *Browse) SetFile(req *http.Request) (err error) {
	f, err := GetString(req, "file")
	if err == nil {
		b.File = f
	}
	return
}

func (b *Browse) SetLevel(route string) {
	switch route {
	case "homeView":
		b.Level = HOME
	case "projectResult":
		b.Level = PROJECTS
	case "userResult":
		b.Level = USERS
	}
}

func (l Level) Home() bool {
	return l == HOME
}

func (l Level) Projects() bool {
	return l == PROJECTS
}

func (l Level) Users() bool {
	return l == USERS
}

func (l Level) Submissions() bool {
	return l == SUBMISSIONS
}

func (l Level) Files() bool {
	return l == FILES
}

func (l Level) Analysis() bool {
	return l == ANALYSIS
}

func (l Level) Chart() bool {
	return l == CHART
}
