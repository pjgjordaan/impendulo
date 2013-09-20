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
	OUT Perm = iota
	ALL
	IN
)

var (
	posters     map[string]Poster
	viewRoutes  map[string]string
	permissions map[string]Perm
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
		"homeView": "home", "skeletonView": "submit",
		"registerView": "register", "projectDownloadView": "download",
		"projectDeleteView": "delete", "userDeleteView": "delete",
		"userResult": "home", "projectResult": "home",
		"archiveView": "submit", "projectView": "submit",
		"statusView": "status", "runToolView": "tool",
	}
}

//Permissions loads all permissions.
func Permissions() map[string]Perm {
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
func defaultPermissions() map[string]Perm {
	return map[string]Perm{
		"homeview": ALL, "skeletonview": IN, "configview": IN,
		"registerview": OUT, "projectdownloadview": IN,
		"projectdeleteview": IN, "userdeleteview": IN, "userresult": ALL,
		"projectresult": ALL, "archiveview": IN, "projectview": IN,
		"addproject": IN, "changeskeleton": IN, "submitarchive": IN,
		"login": OUT, "register": OUT, "logout": IN, "deleteproject": IN,
		"deleteuser": IN, "displaygraph": ALL, "displayresult": ALL, "getfiles": ALL,
		"getsubmissions": ALL, "skeleton.zip": ALL, "index": ALL, "favicon.ico": ALL,
		"": ALL, "statusview": IN, "runtoolview": IN, "runtool": IN,
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
		args := map[string]interface{}{"ctx": ctx}
		return T(getNav(ctx), name).Execute(w, args)
	}
}

//CheckAccess verifies that a user is allowed access to a url.
func CheckAccess(url *url.URL, ctx *Context) error {
	//Rertieve the location they are requesting
	start := strings.LastIndex(url.Path, "/") + 1
	end := strings.Index(url.Path, "?")
	if end < 0 {
		end = len(url.Path)
	}
	if start > end {
		return fmt.Errorf("Invalid request %s", url.Path)
	}
	name := url.Path[start:end]
	perms := Permissions()
	//Get the permission and check it.
	val, ok := perms[name]
	if !ok {
		return fmt.Errorf("Could not find request %s", url.Path)
	}
	switch val {
	case OUT:
		if ctx.LoggedIn() {
			return fmt.Errorf("Cannot access %s when logged in.", url.Path)
		}
	case IN:
		if !ctx.LoggedIn() {
			return fmt.Errorf("Insufficient permissions to access %s", url.Path)
		}
	}
	return nil
}
