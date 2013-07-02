package context

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/httpbuf"
	"net/http"
	"encoding/gob"
)

func init(){
	gob.Register(new(BrowseData))
}

type Context struct {
	Session  *sessions.Session
	projects []*project.Project
	users    []*user.User
	listeners []*jpf.Listener
	Browse *BrowseData
}

type BrowseData struct{
	IsUser bool
	Uid string
	Pid string
	Sid string
}

func (ctx *Context) Close() {
	ctx.save()
}

func (ctx *Context) save(){
	ctx.Session.Values["browse"] = ctx.Browse
}


func (ctx *Context) Save(req *http.Request, buff *httpbuf.Buffer)(error){
	ctx.save()
	return ctx.Session.Save(req, buff)
}

func (ctx *Context) LoggedIn() bool {
	_, err := ctx.Username()
	return err == nil
}

func (ctx *Context) AddMessage(msg string, isErr bool) {
	var tipe string
	if isErr {
		tipe = "error"
	} else {
		tipe = "success"
	}
	ctx.Session.AddFlash(msg, tipe)
}

func (ctx *Context) Errors() []interface{} {
	return ctx.Session.Flashes("error")
}

func (ctx *Context) Successes() []interface{} {
	return ctx.Session.Flashes("success")
}

func (ctx *Context) Username() (string, error) {
	username, ok := ctx.Session.Values["user"].(string)
	if !ok {
		return "", fmt.Errorf("Could not retrieve user.")
	}
	return username, nil
}

func (ctx *Context) AddUser(user string) {
	ctx.Session.Values["user"] = user
}

func (ctx *Context) Projects() ([]*project.Project, error) {
	var err error
	if ctx.projects == nil {
		ctx.projects, err = db.GetProjects(nil, nil)
	}
	return ctx.projects, err
}

func (ctx *Context) Users() ([]*user.User, error) {
	var err error
	if ctx.users == nil {
		ctx.users, err = db.GetUsers(nil)
	}
	return ctx.users, err
}

func (ctx *Context) Listeners() ([]*jpf.Listener, error) {
	var err error
	if ctx.listeners == nil {
		ctx.listeners, err = jpf.Listeners()
	}
	return ctx.listeners, err
}

func NewContext(sess *sessions.Session) *Context {
	ctx := &Context{Session: sess}
	if val, ok := ctx.Session.Values["browse"]; ok{
		ctx.Browse = val.(*BrowseData)
	} else{
		ctx.Browse = new(BrowseData)
	}
	return ctx
}
