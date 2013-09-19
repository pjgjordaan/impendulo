//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
