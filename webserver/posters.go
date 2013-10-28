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
	"code.google.com/p/gorilla/pat"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

type (
	//A function used to fullfill a POST request.
	Poster func(*http.Request, *Context) (string, error)
)

var (
	indexPosters map[string]bool
	posters      map[string]Poster
)

//Posters retrieves all posters
func Posters() map[string]Poster {
	if posters == nil {
		posters = toolPosters()
		defualt := defaultPosters()
		for k, v := range defualt {
			posters[k] = v
		}
	}
	return posters
}

//defaultPosters loads the default posters.
func defaultPosters() map[string]Poster {
	return map[string]Poster{
		"addproject": AddProject, "changeskeleton": ChangeSkeleton,
		"submitarchive": SubmitArchive, "runtool": RunTool,
		"deleteproject": DeleteProject, "deleteuser": DeleteUser,
		"importdata": ImportData, "exportdata": ExportData,
		"evaluatesubmissions": EvaluateSubmissions,
		"login":               Login, "register": Register,
		"logout": Logout, "editproject": EditProject,
		"edituser": EditUser, "editsubmission": EditSubmission,
		"editfile": EditFile,
	}
}

//indexPosters loads the posters which need to be redirected to the home page on success.
func IndexPosters() map[string]bool {
	if indexPosters == nil {
		indexPosters = map[string]bool{
			"login": true, "register": true,
			"logout": true,
		}
	}
	return indexPosters
}

//GeneratePosts loads post request handlers and adds them to the router.
func GeneratePosts(router *pat.Router, posts map[string]Poster, indexPosts map[string]bool) {
	for name, fn := range posts {
		handleFunc := fn.CreatePost(indexPosts[name])
		pattern := "/" + name
		router.Add("POST", pattern, Handler(handleFunc)).Name(name)
	}
}

//CreatePost loads a post request handler.
func (this Poster) CreatePost(indexDest bool) Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		msg, err := this(req, ctx)
		ctx.AddMessage(msg, err != nil)
		if err == nil && indexDest {
			http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
		} else {
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		}
		return err
	}
}

//SubmitArchive adds an Intlola archive to the database.
func SubmitArchive(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	username, msg, err := getActiveUser(ctx)
	if err != nil {
		return
	}
	_, archiveBytes, err := ReadFormFile(req, "archive")
	if err != nil {
		msg = "Could not read archive."
		return
	}
	//We need to create a submission for this archive so that
	//it can be added to the db and so that it can be processed
	sub := project.NewSubmission(projectId, username, project.ARCHIVE_MODE,
		util.CurMilis())
	err = db.Add(db.SUBMISSIONS, sub)
	if err != nil {
		msg = "Could not create submission."
		return
	}
	file := project.NewArchive(sub.Id, archiveBytes)
	err = db.Add(db.FILES, file)
	if err != nil {
		msg = "Could not store archive."
		return
	}
	//Start a submission and send the file to be processed.
	err = processing.StartSubmission(sub.Id)
	if err != nil {
		msg = "Could not start archive submission."
		return
	}
	err = processing.AddFile(file)
	if err != nil {
		msg = "Could not start archive submission."
		return
	}
	err = processing.EndSubmission(sub.Id)
	if err != nil {
		msg = "Could not complete archive submission."
	} else {
		msg = "Archive submitted successfully."
	}
	return
}

//ChangeSkeleton replaces a project's skeleton file.
func ChangeSkeleton(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	_, data, err := ReadFormFile(req, "skeleton")
	if err != nil {
		msg = "Could not read skeleton file."
		return
	}
	err = db.Update(
		db.PROJECTS, bson.M{db.ID: projectId},
		bson.M{db.SET: bson.M{db.SKELETON: data}},
	)
	if err != nil {
		msg = "Could not update skeleton file."
	} else {
		msg = "Successfully updated skeleton file."
	}
	return
}

//AddProject creates a new Impendulo Project.
func AddProject(req *http.Request, ctx *Context) (msg string, err error) {
	name, err := GetString(req, "projectname")
	if err != nil {
		msg = "Could not read project name."
		return
	}
	lang, err := GetString(req, "lang")
	if err != nil {
		msg = "Could not read project language."
		return
	}
	username, msg, err := getActiveUser(ctx)
	if err != nil {
		return
	}
	_, skeletonBytes, err := ReadFormFile(req, "skeleton")
	if err != nil {
		msg = "Could not read skeleton file."
		return
	}
	p := project.NewProject(name, username, lang, skeletonBytes)
	err = db.Add(db.PROJECTS, p)
	if err != nil {
		msg = "Could not add project."
	} else {
		msg = "Successfully added project."
	}
	return
}

//DeleteProject removes a project and all data associated with it from the system.
func DeleteProject(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	err = db.RemoveProjectById(projectId)
	if err != nil {
		msg = "Could not delete project."
	} else {
		msg = "Successfully deleted project."
	}
	return
}

//EditProject
func EditProject(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	name, err := GetString(req, "projectname")
	if err != nil {
		msg = "Could not read project name."
		return
	}
	u, msg, err := getUserId(req)
	if err != nil {
		return
	}
	if !db.Contains(db.USERS, bson.M{db.ID: u}) {
		err = fmt.Errorf("Invalid user %s.", u)
		msg = err.Error()
		return
	}
	lang, err := GetString(req, "lang")
	if err != nil {
		msg = "Could not read language."
		return
	}
	if !tool.Supported(lang) {
		err = fmt.Errorf("Unsupported language %s.", lang)
		msg = err.Error()
		return
	}
	change := bson.M{db.SET: bson.M{db.NAME: name, db.USER: u, db.LANG: lang}}
	err = db.Update(db.PROJECTS, bson.M{db.ID: projectId}, change)
	if err != nil {
		msg = "Could not edit project."
	} else {
		msg = "Successfully edited project."
	}
	return
}

//EditSubmission
func EditSubmission(req *http.Request, ctx *Context) (msg string, err error) {
	subId, msg, err := getSubId(req)
	if err != nil {
		return
	}
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	if !db.Contains(db.PROJECTS, bson.M{db.ID: projectId}) {
		err = fmt.Errorf("Invalid project %s.", projectId.Hex())
		msg = err.Error()
		return
	}
	u, msg, err := getUserId(req)
	if err != nil {
		return
	}
	if !db.Contains(db.USERS, bson.M{db.ID: u}) {
		err = fmt.Errorf("Invalid user %s.", u)
		msg = err.Error()
		return
	}
	change := bson.M{db.SET: bson.M{db.PROJECTID: projectId, db.USER: u}}
	err = db.Update(db.SUBMISSIONS, bson.M{db.ID: subId}, change)
	if err != nil {
		msg = "Could not edit submission."
	} else {
		msg = "Successfully edited submission."
	}
	return
}

//EditFile
func EditFile(req *http.Request, ctx *Context) (msg string, err error) {
	fileId, msg, err := getFileId(req)
	if err != nil {
		return
	}
	subId, msg, err := getSubId(req)
	if err != nil {
		return
	}
	if !db.Contains(db.SUBMISSIONS, bson.M{db.ID: subId}) {
		err = fmt.Errorf("Invalid submission %s.", subId.Hex())
		msg = err.Error()
		return
	}
	name, err := GetString(req, "filename")
	if err != nil {
		msg = "Could not read file name."
		return
	}
	pkg, err := GetString(req, "package")
	if err != nil {
		msg = "Could not read package."
		return
	}
	change := bson.M{db.SET: bson.M{db.SUBID: subId, db.NAME: name, db.PKG: pkg}}
	err = db.Update(db.FILES, bson.M{db.ID: fileId}, change)
	if err != nil {
		msg = "Could not edit file."
	} else {
		msg = "Successfully edited file."
	}
	return
}

func EvaluateSubmissions(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	all := req.FormValue("projectid") == "all"
	if err != nil && !all {
		return
	}
	var matcher bson.M
	if all {
		matcher = bson.M{}
	} else {
		matcher = bson.M{db.PROJECTID: projectId}
	}
	submissions, err := db.Submissions(matcher, nil)
	if err != nil {
		msg = "Could not retrieve submissions."
		return
	}
	for _, submission := range submissions {
		err = db.UpdateStatus(submission)
		if err != nil {
			msg = fmt.Sprintf("Could not evaluate submission %s.", submission.Id.Hex())
			return
		}
		err = db.UpdateTime(submission)
		if err != nil {
			msg = fmt.Sprintf("Could not evaluate submission %s.", submission.Id.Hex())
			return
		}
	}
	msg = "Successfully evaluated submissions."
	return
}

//Login signs a user into the web app.
func Login(req *http.Request, ctx *Context) (msg string, err error) {
	uname, pword, msg, err := getCredentials(req)
	if err != nil {
		return
	}
	u, err := db.User(uname)
	if err != nil {
		msg = fmt.Sprintf("User %s not found.", uname)
		return
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("Invalid username or password.")
		msg = err.Error()
	} else {
		ctx.AddUser(uname)
		msg = "Logged in successfully."
	}
	return
}

//Register registers a new user with Impendulo.
func Register(req *http.Request, ctx *Context) (msg string, err error) {
	uname, pword, msg, err := getCredentials(req)
	if err != nil {
		return
	}
	u := user.New(uname, pword)
	err = db.Add(db.USERS, u)
	if err != nil {
		msg = fmt.Sprintf("User %s already exists.", uname)
	} else {
		ctx.AddUser(uname)
		msg = "Registered successfully."
	}
	return
}

//DeleteUser removes a user and all data associated with them from the system.
func DeleteUser(req *http.Request, ctx *Context) (msg string, err error) {
	uname, msg, err := getString(req, "username")
	if err != nil {
		return
	}
	err = db.RemoveUserById(uname)
	if err != nil {
		msg = fmt.Sprintf("Could not delete user %s.", uname)
	} else {
		msg = fmt.Sprintf("Successfully deleted user %s.", uname)
	}
	return
}

//Logout logs a user out of the system.
func Logout(req *http.Request, ctx *Context) (string, error) {
	ctx.RemoveUser()
	return "Successfully logged out.", nil
}

//EditUser
func EditUser(req *http.Request, ctx *Context) (msg string, err error) {
	oldName, err := GetString(req, "oldname")
	if err != nil {
		msg = "Could not read old username."
		return
	}
	newName, err := GetString(req, "newname")
	if err != nil {
		msg = "Could not read new username."
		return
	}
	access, err := GetInt(req, "access")
	if err != nil {
		msg = "Could not read user access level."
		return
	}
	if !user.ValidPermission(access) {
		err = fmt.Errorf("Invalid user access level %d.", access)
		msg = err.Error()
		return
	}
	if oldName != newName {
		err = db.RenameUser(oldName, newName)
		if err != nil {
			msg = fmt.Sprintf("Could not rename user %s to %s.", oldName, newName)
			return
		}
	}
	change := bson.M{db.SET: bson.M{user.ACCESS: access}}
	err = db.Update(db.USERS, bson.M{db.ID: newName}, change)
	if err != nil {
		msg = "Could not edit user."
	} else {
		msg = "Successfully edited user."
	}
	//Ugly hack should change this.
	current := req.Header.Get("Referer")
	current = current[:strings.LastIndex(current, "/")+1] + "editdbview?editing=User"
	req.Header.Set("Referer", current)
	return
}
