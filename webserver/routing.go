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
	"net/http"
	"strings"
)

var (
	posters     map[string]Poster
	viewRoutes  map[string]string
	permissions map[string]int
)

type (
	//A function used to add data to the database.
	Poster func(*http.Request, *Context) (string, error)
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

//Posters
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

//defaultPosters
func defaultPosters() map[string]Poster {
	return map[string]Poster{
		"addproject": AddProject, "changeskeleton": ChangeSkeleton,
		"submitarchive": SubmitArchive, "runtool": RunTool,
		"deleteproject": DeleteProject, "deleteuser": DeleteUser,
	}
}

//GeneratePosts loads post request handlers.
func GeneratePosts(router *pat.Router) {
	posts := Posters()
	for name, fn := range posts {
		handleFunc := fn.CreatePost()
		pattern := "/" + name
		router.Add("POST", pattern, Handler(handleFunc)).Name(name)
	}
}

//Views
func Views() map[string]string {
	if viewRoutes != nil {
		return viewRoutes
	}
	viewRoutes = defaultViews()
	return viewRoutes
}

//defaultViews
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

//Permissions
func Permissions() map[string]int {
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

//defaultPermissions
func defaultPermissions() map[string]int {
	return map[string]int{
		"homeview": 0, "skeletonview": 1, "configview": 1,
		"registerview": 0, "projectdownloadview": 1,
		"projectdeleteview": 1, "userdeleteview": 1, "userresult": 0,
		"projectresult": 0, "archiveview": 1, "projectview": 1,
		"addproject": 1, "changeskeleton": 1, "submitarchive": 1,
		"login": 0, "register": 0, "logout": 1, "deleteproject": 1,
		"deleteuser": 1, "displaygraph": 0, "displayresult": 0, "getfiles": 0,
		"getsubmissions": 0, "skeleton.zip": 0, "index": 0, "favicon.ico": 0,
		"": 0, "statusview": 1, "runtoolview": 1, "runtool": 1,
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
