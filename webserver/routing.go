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
	"github.com/godfried/impendulo/user"
	"net/http"
	"net/url"
	"strings"
)

type (
	//A function used to add data to the database.
	Poster func(*http.Request, *Context) (string, error)
	Perm   int
)

const (
	OUT user.Permission = 42
)

var (
	posters     map[string]Poster
	viewRoutes  map[string]string
	permissions map[string]user.Permission
)

//CreatePost loads a post request handler.
func (this Poster) CreatePost() Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		msg, err := this(req, ctx)
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
}

//Posters retrieves all posters
func Posters() map[string]Poster {
	if posters != nil {
		return posters
	}
	posters = toolPosters()
	defualt := defaultPosters()
	for k, v := range defualt {
		posters[k] = v
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
	}
}

//GeneratePosts loads post request handlers and adds them to the router.
func GeneratePosts(router *pat.Router) {
	posts := Posters()
	for name, fn := range posts {
		handleFunc := fn.CreatePost()
		pattern := "/" + name
		router.Add("POST", pattern, Handler(handleFunc)).Name(name)
	}
}

//Views loads all views.
func Views() map[string]string {
	if viewRoutes != nil {
		return viewRoutes
	}
	viewRoutes = defaultViews()
	return viewRoutes
}

//defaultViews loads default views.
func defaultViews() map[string]string {
	return map[string]string{
		"homeView": "home", "userResult": "home", "projectResult": "home",
		"skeletonView": "submit", "archiveView": "submit", "projectView": "submit",
		"registerView":        "register",
		"projectDownloadView": "download",
		"projectDeleteView":   "delete", "userDeleteView": "delete",
		"statusView": "status", "runToolView": "tool",
		"importDataView": "data", "exportDataView": "data",
	}
}

//Permissions loads all permissions.
func Permissions() map[string]user.Permission {
	if permissions != nil {
		return permissions
	}
	permissions = toolPermissions()
	defualt := defaultPermissions()
	for k, v := range defualt {
		permissions[k] = v
	}
	return permissions
}

//defaultPermissions loads default permissions.
func defaultPermissions() map[string]user.Permission {
	return map[string]user.Permission{
		"registerview": OUT, "register": OUT, "login": OUT,
		"homeview": user.NONE, "projectresult": user.NONE,
		"userresult": user.NONE, "": user.NONE, "displaychart": user.NONE,
		"displayresult": user.NONE, "getfiles": user.NONE,
		"index": user.NONE, "favicon.ico": user.NONE, "getsubmissions": user.NONE,
		"projectdownloadview": user.STUDENT, "skeleton.zip": user.STUDENT,
		"archiveview": user.STUDENT, "submitarchive": user.STUDENT,
		"logout":       user.STUDENT,
		"skeletonview": user.TEACHER, "changeskeleton": user.TEACHER,
		"projectview": user.TEACHER, "addproject": user.TEACHER,
		"runtoolview": user.TEACHER, "runtool": user.TEACHER,
		"configview":        user.TEACHER,
		"projectdeleteview": user.ADMIN, "deleteproject": user.ADMIN,
		"userdeleteview": user.ADMIN, "deleteuser": user.ADMIN,
		"importdataview": user.ADMIN, "exportdataview": user.ADMIN,
		"importdata": user.ADMIN, "exportdata": user.ADMIN,
		"statusview": user.ADMIN,
	}
}

//GenerateViews is used to load all the basic views used by our web app.
func GenerateViews(router *pat.Router) {
	views := Views()
	for name, view := range views {
		handleFunc := LoadView(name, view)
		lname := strings.ToLower(name)
		pattern := "/" + lname
		router.Add("GET", pattern, Handler(handleFunc)).Name(lname)
	}
}

//LoadView loads a view so that it is accessible in our web app.
func LoadView(name, view string) Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		views := Views()
		ctx.Browse.View = views[name]
		if ctx.Browse.View == "home" {
			ctx.Browse.SetLevel(name)
		}
		args := map[string]interface{}{"ctx": ctx}
		return T(getNav(ctx), name).Execute(w, args)
	}
}

//CheckAccess verifies that a user is allowed access to a url.
func CheckAccess(url *url.URL, ctx *Context) (err error) {
	//Rertieve the location they are requesting
	start := strings.LastIndex(url.Path, "/") + 1
	end := strings.Index(url.Path, "?")
	if end < 0 {
		end = len(url.Path)
	}
	if start > end {
		err = fmt.Errorf("Invalid request %s", url.Path)
		return
	}
	name := url.Path[start:end]
	perms := Permissions()
	//Get the permission and check it.
	val, ok := perms[name]
	if !ok {
		err = fmt.Errorf("Could not find request %s", url.Path)
		return
	}
	var msg string
	//Check permission levels.
	switch val {
	case OUT:
		if ctx.LoggedIn() {
			msg = "Cannot access %s when logged in."
		}
	case user.STUDENT:
		if !ctx.LoggedIn() {
			msg = "You need to be logged in to access %s"
		}
	case user.ADMIN, user.TEACHER:
		if !ctx.LoggedIn() {
			msg = "You need to be logged in to access %s"
		} else {
			uname, _ := ctx.Username()
			if !checkUserPermission(uname, val) {
				msg = "You have insufficient permissions to access %s"
			}
		}
	}
	if msg != "" {
		err = fmt.Errorf(msg, url.Path)
	}
	return
}

//checkUserPermission verifies that a user has the specified permission level.
func checkUserPermission(uname string, perm user.Permission) bool {
	user, err := db.User(uname)
	if err != nil {
		return false
	}
	return user.Access >= perm
}
