package context

import (
	"code.google.com/p/gorilla/sessions"
	"net/http"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type Context struct {
	Session  *sessions.Session
	Projects []*project.Project
	Users []*user.User
	ProjectSubs map[bson.ObjectId][]*project.Submission
	UserSubs map[string][]*project.Submission
	Files map[bson.ObjectId][]*project.File
}

func (c *Context) Close() { 
	db.Close() 
}

func (ctx *Context) LoadProjects() error {
	var err error
	ctx.Projects, err = db.GetProjects(nil, nil)
	return err
}


func (ctx *Context) LoadUsers() error {
	var err error
	ctx.Users, err = db.GetUsers(nil)
	return err
}

func (ctx *Context) Subs(tipe, idStr string)([]*project.Submission, error){
	if tipe == "project"{
		return ctx.loadProjectSubs(idStr)
	} else if tipe == "user"{
		return ctx.loadUserSubs(idStr)
	}
	return nil, fmt.Errorf("Unknown request type %q", tipe)
}

func (ctx *Context) loadUserSubs(user string)([]*project.Submission, error){
	if ctx.UserSubs[user] == nil{
		var err error
		ctx.UserSubs[user], err = db.GetSubmissions(bson.M{project.USER:user}, nil)
		if err != nil {
			return nil, err
		}
	}
	return ctx.UserSubs[user], nil
}


func (ctx *Context) loadProjectSubs(idStr string)([]*project.Submission, error){
	if !bson.IsObjectIdHex(idStr){
		return nil, fmt.Errorf("Invalid id %q", idStr)
	}
	pid := bson.ObjectIdHex(idStr)
	if ctx.ProjectSubs[pid] == nil{
		var err error
		ctx.ProjectSubs[pid], err = db.GetSubmissions(bson.M{project.PROJECT_ID:pid}, nil)
		if err != nil {
			return nil, err
		}
	}
	return ctx.ProjectSubs[pid], nil
}

func (ctx *Context) GetFiles(idStr string)([]*project.File, error){
	if !bson.IsObjectIdHex(idStr){
		return nil, fmt.Errorf("Invalid submission id %q", idStr)
	}
	subId := bson.ObjectIdHex(idStr)
	if ctx.Files[subId] == nil{
		var err error
		choices := []bson.M{bson.M{project.INFO+"."+project.TYPE:project.EXEC}, bson.M{project.INFO+"."+project.TYPE:project.SRC}}
		ctx.Files[subId], err = db.GetFiles(bson.M{project.SUBID: subId, "$or": choices}, bson.M{project.INFO: 1})
		if err != nil {
			return nil, err
		}
	}
	return ctx.Files[subId], nil
}


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

func (ctx *Context) AddError(err string){
	ctx.Session.AddFlash(err, "error")
}


func (ctx *Context) AddSuccess(msg string){
	ctx.Session.AddFlash(msg, "success")
}


func (ctx *Context) Errors()[]interface{}{
	return ctx.Session.Flashes("error")
}


func (ctx *Context) Successes()[]interface{}{
	return ctx.Session.Flashes("success")
}

func NewContext(req *http.Request, sess *sessions.Session) *Context {
	db.Setup(db.DEFAULT_CONN)
	ctx := &Context{Session:  sess}
	//ctx.loadProjects()
	ctx.ProjectSubs = make(map[bson.ObjectId][]*project.Submission)
	ctx.UserSubs = make(map[string][]*project.Submission)
	ctx.Files = make(map[bson.ObjectId][]*project.File)
	return ctx
}
