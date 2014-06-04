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

package web

import (
	"code.google.com/p/gorilla/pat"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/user"

	"net/http"
	"strings"
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
		"userresult", "displayresult",
		"getfiles", "favicon.ico", "getsubmissions", "submissionschartview",
		"static", "userchart", "projectchart",
	}
	student = []string{
		"testdownloadview", "test.zip",
		"projectdownloadview", "skeleton.zip",
		"intloladownloadview", "intlola.zip",
		"archiveview", "submitarchive", "logout",
	}
	teacher = []string{
		"skeletonview", "addskeleton", "projectview",
		"addproject", "runtoolsview", "runtools", "configview",
	}
	admin = []string{
		"deleteprojects", "deleteusers", "deleteresults", "deleteview",
		"deleteskeletons", "deletesubmissions", "importdataview", "exportdataview",
		"importdata", "exportdb.zip", "statusview", "evaluatesubmissionsview",
		"evaluatesubmissions", "logs", "editdbview", "loadproject", "editproject",
		"loaduser", "edituser", "loadsubmission", "editsubmission", "loadfile",
		"editfile", "edittest", "renamefiles", "renameview",
	}

	homeViews = []string{
		"homeview", "userresult", "projectresult",
		"userchart", "projectchart",
		"displayresult", "getfiles", "submissionschartview",
		"getsubmissions",
	}
	submitViews = []string{
		"skeletonview", "archiveview", "projectview",
		"configview",
	}
	registerViews = []string{"registerview"}
	downloadViews = []string{"projectdownloadview", "intloladownloadview", "testdownloadview"}
	statusViews   = []string{"statusview"}
	toolViews     = []string{"runtoolsview", "evaluatesubmissionsview"}
	dataViews     = []string{
		"importdataview", "exportdataview", "editdbview", "renameview",
		"loadproject", "loadsubmission", "loadfile", "loaduser", "deleteview",
	}
)

//Views loads all views.
func Views() map[string]string {
	if viewRoutes != nil {
		return viewRoutes
	}
	viewRoutes = make(map[string]string)
	setViewRoutes(homeViews, "home")
	setViewRoutes(submitViews, "submit")
	setViewRoutes(registerViews, "register")
	setViewRoutes(downloadViews, "download")
	setViewRoutes(statusViews, "status")
	setViewRoutes(toolViews, "tool")
	setViewRoutes(dataViews, "data")
	return viewRoutes
}

func setViewRoutes(routes []string, view string) {
	for _, r := range routes {
		viewRoutes[r] = view
	}
}

//Permissions loads all permissions.
func Permissions() map[string]user.Permission {
	if permissions != nil {
		return permissions
	}
	permissions = toolPermissions()
	setPermissions(none, user.NONE)
	setPermissions(out, OUT)
	setPermissions(student, user.STUDENT)
	setPermissions(teacher, user.TEACHER)
	setPermissions(admin, user.ADMIN)
	return permissions
}

func setPermissions(views []string, p user.Permission) {
	for _, v := range views {
		permissions[v] = p
	}
}

//GenerateViews is used to load all the basic views used by our web app.
func GenerateViews(r *pat.Router, views map[string]string) {
	for n, v := range views {
		r.Add("GET", "/"+n, Handler(LoadView(n, v))).Name(n)
	}
}

//LoadView loads a view so that it is accessible in our web app.
func LoadView(n, v string) Handler {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		c.Browse.View = v
		return T(getNav(c), n).Execute(w, map[string]interface{}{"ctx": c})
	}
}

//CheckAccess verifies that a user is allowed access to a url.
func CheckAccess(p string, c *Context, ps map[string]user.Permission) error {
	//Retrieve the location they are requesting
	n := p
	if strings.HasPrefix(n, "/") {
		if len(n) > 1 {
			n = n[1:]
		} else {
			n = ""
		}
	}
	if i := strings.Index(n, "/"); i != -1 {
		n = n[:i]
	}
	if i := strings.Index(n, "?"); i != -1 {
		n = n[:i]
	}
	//Get the permission and check it.
	v, ok := ps[n]
	if !ok {
		return fmt.Errorf("could not find request %s", n)
	}
	return checkPermission(c, v, p)
}

func checkPermission(c *Context, p user.Permission, url string) error {
	//Check permission levels.
	switch p {
	case user.NONE:
	case OUT:
		if c.LoggedIn() {
			return fmt.Errorf("cannot access %s when logged in", url)
		}
	case user.STUDENT:
		if !c.LoggedIn() {
			return fmt.Errorf("login required to access %s", url)
		}
	case user.ADMIN, user.TEACHER:
		u, e := c.Username()
		if e != nil {
			return fmt.Errorf("login required to access %s", url)
		} else if !checkUserPermission(u, p) {
			return fmt.Errorf("insufficient permissions to access %s", url)
		}
	default:
		return fmt.Errorf("unknown url %s", url)
	}
	return nil
}

//checkUserPermission verifies that a user has the specified permission level.
func checkUserPermission(id string, p user.Permission) bool {
	u, e := db.User(id)
	return e == nil && u.Access >= p
}
