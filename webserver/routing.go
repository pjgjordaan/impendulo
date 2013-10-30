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
	"strings"
)

type (
	Perm int
)

const (
	OUT user.Permission = 42
)

var (
	viewRoutes  map[string]string
	permissions map[string]user.Permission
	out         = []string{
		"registerview", "register", "login",
	}
	none = []string{
		"index", "", "homeview", "projectresult",
		"userresult", "displaychart", "displayresult", "getfiles",
		"favicon.ico", "getsubmissions", "getsubmissionschart",
		"static", "userchart", "projectchart",
	}
	student = []string{
		"projectdownloadview", "skeleton.zip",
		"archiveview", "submitarchive", "logout",
	}
	teacher = []string{
		"skeletonview", "changeskeleton", "projectview",
		"addproject", "runtoolview", "runtool", "configview",
	}
	admin = []string{
		"projectdeleteview", "deleteproject", "userdeleteview",
		"deleteuser", "resultsdeleteview", "deleteresults",
		"importdataview", "exportdataview",
		"importdata", "exportdb.zip", "statusview",
		"evaluatesubmissionsview", "evaluatesubmissions", "logs",
		"editdbview", "loadproject", "editproject", "loaduser",
		"edituser", "loadsubmission", "editsubmission", "loadfile",
		"editfile",
	}

	homeViews = []string{
		"homeview", "userresult", "projectresult",
		"userchart", "projectchart", "displaychart",
		"displayresult", "getfiles", "getsubmissionschart",
		"getsubmissions",
	}
	submitViews = []string{
		"skeletonview", "archiveview", "projectview",
		"configview",
	}
	registerViews = []string{"registerview"}
	downloadViews = []string{"projectdownloadview"}
	deleteViews   = []string{"projectdeleteview", "userdeleteview", "resultsdeleteview"}
	statusViews   = []string{"statusview"}
	toolViews     = []string{"runtoolview", "evaluatesubmissionsview"}
	dataViews     = []string{
		"importdataview", "exportdataview", "editdbview",
		"loadproject", "loadsubmission", "loadfile", "loaduser",
	}
)

//Views loads all views.
func Views() map[string]string {
	if viewRoutes != nil {
		return viewRoutes
	}
	viewRoutes = make(map[string]string)
	for _, name := range homeViews {
		viewRoutes[name] = "home"
	}
	for _, name := range submitViews {
		viewRoutes[name] = "submit"
	}
	for _, name := range registerViews {
		viewRoutes[name] = "register"
	}
	for _, name := range downloadViews {
		viewRoutes[name] = "download"
	}
	for _, name := range deleteViews {
		viewRoutes[name] = "delete"
	}
	for _, name := range statusViews {
		viewRoutes[name] = "status"
	}
	for _, name := range toolViews {
		viewRoutes[name] = "tool"
	}
	for _, name := range dataViews {
		viewRoutes[name] = "data"
	}
	return viewRoutes
}

//Permissions loads all permissions.
func Permissions() map[string]user.Permission {
	if permissions != nil {
		return permissions
	}
	permissions = toolPermissions()
	for _, name := range none {
		permissions[name] = user.NONE
	}
	for _, name := range out {
		permissions[name] = OUT
	}
	for _, name := range student {
		permissions[name] = user.STUDENT
	}
	for _, name := range teacher {
		permissions[name] = user.TEACHER
	}
	for _, name := range admin {
		permissions[name] = user.ADMIN
	}
	return permissions
}

//GenerateViews is used to load all the basic views used by our web app.
func GenerateViews(router *pat.Router, views map[string]string) {
	for name, view := range views {
		handleFunc := LoadView(name, view)
		pattern := "/" + name
		router.Add("GET", pattern, Handler(handleFunc)).Name(name)
	}
}

//LoadView loads a view so that it is accessible in our web app.
func LoadView(name, view string) Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		ctx.Browse.View = view
		if ctx.Browse.View == "home" {
			ctx.Browse.SetLevel(name)
		}
		args := map[string]interface{}{"ctx": ctx}
		return T(getNav(ctx), name).Execute(w, args)
	}
}

//CheckAccess verifies that a user is allowed access to a url.
func CheckAccess(path string, ctx *Context, perms map[string]user.Permission) (err error) {
	//Retrieve the location they are requesting
	name := path
	if strings.HasPrefix(name, "/") {
		if len(name) > 1 {
			name = name[1:]
		} else {
			name = ""
		}
	}
	if index := strings.Index(name, "/"); index != -1 {
		name = name[:index]
	}
	if index := strings.Index(name, "?"); index != -1 {
		name = name[:index]
	}
	//Get the permission and check it.
	val, ok := perms[name]
	if !ok {
		err = fmt.Errorf("Could not find request %s", name)
		return
	}
	if msg := checkPermission(ctx, val); msg != "" {
		err = fmt.Errorf(msg, path)
	}
	return
}

func checkPermission(ctx *Context, perm user.Permission) (msg string) {
	//Check permission levels.
	loggedIn := ctx.LoggedIn()
	switch perm {
	case user.NONE:
	case OUT:
		if loggedIn {
			msg = "Cannot access %s when logged in."
		}
	case user.STUDENT:
		if !loggedIn {
			msg = "You need to be logged in to access %s"
		}
	case user.ADMIN, user.TEACHER:
		uname, err := ctx.Username()
		if err != nil {
			msg = "You need to be logged in to access %s"
		} else if !checkUserPermission(uname, perm) {
			msg = "You have insufficient permissions to access %s"
		}
	default:
		msg = "Unknown url %s"
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
