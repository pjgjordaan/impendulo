package context

import (
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/db"
	"fmt"
)

type Context struct {
	Session  *sessions.Session
	projects []*project.Project
	users []*user.User
}

func (ctx *Context) Close() { }

func (ctx *Context) LoggedIn() bool{
	_, err := ctx.Username()
	return err == nil
}

func (ctx *Context) Username()(string, error){
	username, ok := ctx.Session.Values["user"].(string)
	if !ok{
		return "", fmt.Errorf("Could not retrieve user.")
	}
	return username, nil
}

func (ctx *Context) AddMessage(msg string, isErr bool){
	var tipe string
	if isErr{
		tipe = "error"
	} else{
		tipe = "success"
	}
	ctx.Session.AddFlash(msg, tipe)
}

func (ctx *Context) Errors()[]interface{}{
	return ctx.Session.Flashes("error")
}


func (ctx *Context) Successes()[]interface{}{
	return ctx.Session.Flashes("success")
}

func (ctx *Context) AddUser(user string){
	ctx.Session.Values["user"] = user
}

func (ctx *Context) Projects() ([]*project.Project, error) {
	var err error
	if ctx.projects == nil{
		ctx.projects, err = db.GetProjects(nil, nil)
	}
	return ctx.projects, err
}

func (ctx *Context) Users()([]*user.User, error) {
	var err error
	if ctx.users == nil{
		ctx.users, err = db.GetUsers(nil)
	} 
	return ctx.users, err
}


func NewContext(sess *sessions.Session) *Context {
	ctx := &Context{Session:  sess}
	return ctx
}
