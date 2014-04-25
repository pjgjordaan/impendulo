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
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/mongo"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"net/http"
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
		"addproject": AddProject, "addskeleton": AddSkeleton,
		"submitarchive": SubmitArchive, "runtools": RunTools,
		"deleteproject": DeleteProject, "deleteuser": DeleteUser,
		"deleteresults": DeleteResults, "deleteskeletons": DeleteSkeletons,
		"importdata":          ImportData,
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
	for n, f := range posts {
		h := f.CreatePost(indexPosts[n])
		p := "/" + n
		router.Add("POST", p, Handler(h)).Name(n)
	}
}

//CreatePost loads a post request handler.
func (this Poster) CreatePost(index bool) Handler {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		m, e := this(r, c)
		c.AddMessage(m, e != nil)
		if e == nil && index {
			http.Redirect(w, r, getRoute("index"), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		}
		return e
	}
}

//SubmitArchive adds an Intlola archive to the database.
func SubmitArchive(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	un, m, e := getActiveUser(c)
	if e != nil {
		return m, e
	}
	_, a, e := ReadFormFile(r, "archive")
	if e != nil {
		return "Could not read archive.", e
	}
	//We need to create a submission for this archive so that
	//it can be added to the db and so that it can be processed
	s := project.NewSubmission(pid, un, project.ARCHIVE_MODE, util.CurMilis())
	if e = db.Add(db.SUBMISSIONS, s); e != nil {
		return "Could not create submission.", e
	}
	f := project.NewArchive(s.Id, a)
	if e = db.Add(db.FILES, f); e != nil {
		return "Could not store archive.", e
	}
	k, e := processing.StartSubmission(s.Id)
	if e != nil {
		return "Could not start archive submission.", e
	}
	if e = processing.AddFile(f, k); e != nil {
		return "Could not submit archive.", e
	}
	if e = processing.EndSubmission(s.Id, k); e != nil {
		return "Could not complete archive submission.", e
	}
	return "Archive submitted successfully.", nil
}

//AddProject creates a new Impendulo Project.
func AddProject(r *http.Request, c *Context) (string, error) {
	n, e := GetString(r, "projectname")
	if e != nil {
		return "Could not read project name.", e
	}
	l, e := GetString(r, "lang")
	if e != nil {
		return "Could not read project language.", e
	}
	un, m, e := getActiveUser(c)
	if e != nil {
		return m, e
	}
	if e = db.Add(db.PROJECTS, project.New(n, un, l)); e != nil {
		return "Could not add project.", e
	}
	return "Successfully added project.", nil
}

func AddSkeleton(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	n, e := GetString(r, "skeletonname")
	if e != nil {
		return "Could not read skeleton name.", e
	}
	_, s, e := ReadFormFile(r, "skeleton")
	if e != nil {
		return "Could not read skeleton file.", e
	}
	if e = db.Add(db.SKELETONS, project.NewSkeleton(pid, n, s)); e != nil {
		return "Could not add skeleton.", e
	}
	return "Successfully added skeleton.", nil
}

//DeleteProject removes a project and all data associated with it from the system.
func DeleteProject(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	if e = db.RemoveProjectById(pid); e != nil {
		return "Could not delete project.", e
	}
	return "Successfully deleted project.", nil
}

func DeleteSkeletons(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	ss, e := db.Skeletons(bson.M{db.PROJECTID: pid}, bson.M{db.ID: 1})
	if e != nil {
		return "Could not retrieve skeletons.", e
	}
	for _, s := range ss {
		if e = db.RemoveById(db.SKELETONS, s.Id); e != nil {
			return "Could not delete skeletons.", e
		}
	}
	return "Successfully deleted skeletons.", nil
}

//DeleteResults removes all results for a specic project.
func DeleteResults(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	ss, e := db.Submissions(bson.M{db.PROJECTID: pid}, bson.M{db.ID: 1})
	if e != nil {
		return "Could not retrieve submissions.", e
	}
	for _, s := range ss {
		fs, e := db.Files(bson.M{db.SUBID: s.Id}, bson.M{db.DATA: 0}, 0)
		if e != nil {
			util.Log(e)
			continue
		}
		for _, f := range fs {
			for _, r := range f.Results {
				id, ok := r.(bson.ObjectId)
				if !ok {
					continue
				}
				if e = db.RemoveById(db.RESULTS, id); e != nil {
					util.Log(e)
				}
			}
			if e = db.Update(db.FILES, bson.M{db.ID: f.Id}, bson.M{db.SET: bson.M{db.RESULTS: bson.M{}}}); e != nil {
				util.Log(e)
			}
		}
	}
	return "Successfully deleted results.", nil
}

//EditProject is used to modify a project's metadata.
func EditProject(r *http.Request, c *Context) (string, error) {
	pid, e := util.ReadId(r.FormValue("project-id"))
	if e != nil {
		return "Could not read project id.", e
	}
	n, e := GetString(r, "project-name")
	if e != nil {
		return "Could not read project name.", e
	}
	u, e := GetString(r, "project-user")
	if e != nil {
		return "Could not read user.", e
	}
	if !db.Contains(db.USERS, bson.M{db.ID: u}) {
		e = fmt.Errorf("invalid user %s", u)
		return e.Error(), e
	}
	l, e := GetString(r, "project-lang")
	if e != nil {
		return "Could not read language.", e
	}
	if !tool.Supported(tool.Language(l)) {
		e = fmt.Errorf("unsupported language %s", l)
		return e.Error(), e
	}
	if e = db.Update(db.PROJECTS, bson.M{db.ID: pid}, bson.M{db.SET: bson.M{db.NAME: n, db.USER: u, db.LANG: l}}); e != nil {
		return "Could not edit project.", e
	}
	return "Successfully edited project.", nil
}

//EditSubmission
func EditSubmission(r *http.Request, c *Context) (string, error) {
	sid, e := util.ReadId(r.FormValue("submission-id"))
	if e != nil {
		return "Could not read submission id.", e
	}
	pid, e := util.ReadId(r.FormValue("submission-project"))
	if e != nil {
		return "Could not read project.", e
	}
	if !db.Contains(db.PROJECTS, bson.M{db.ID: pid}) {
		e = fmt.Errorf("invalid project %s", pid.Hex())
		return e.Error(), e
	}
	u, e := GetString(r, "submission-user")
	if e != nil {
		return "Could not read user.", e
	}
	if !db.Contains(db.USERS, bson.M{db.ID: u}) {
		e = fmt.Errorf("invalid user %s", u)
		return e.Error(), e
	}
	if e = db.Update(db.SUBMISSIONS, bson.M{db.ID: sid}, bson.M{db.SET: bson.M{db.PROJECTID: pid, db.USER: u}}); e != nil {
		return "Could not edit submission.", e
	}
	return "Successfully edited submission.", nil
}

//EditFile
func EditFile(r *http.Request, c *Context) (string, error) {
	fid, e := util.ReadId(r.FormValue("file-id"))
	if e != nil {
		return "Could not read file id.", e
	}
	n, e := GetString(r, "file-name")
	if e != nil {
		return "Could not read file name.", e
	}
	p, e := GetString(r, "file-package")
	if e != nil {
		return "Could not read package.", e
	}
	if e = db.Update(db.FILES, bson.M{db.ID: fid}, bson.M{db.SET: bson.M{db.NAME: n, db.PKG: p}}); e != nil {
		return "Could not edit file.", e
	}
	return "Successfully edited file.", nil
}

func EvaluateSubmissions(r *http.Request, c *Context) (string, error) {
	all := r.FormValue("projectid") == "all"
	pid, msg, e := getProjectId(r)
	if e != nil && !all {
		return msg, e
	}
	m := bson.M{}
	if !all {
		m[db.PROJECTID] = pid
	}
	ss, e := db.Submissions(m, nil)
	if e != nil {
		return "Could not retrieve submissions.", e
	}
	for _, s := range ss {
		if e = db.UpdateStatus(s); e != nil {
			return fmt.Sprintf("Could not evaluate submission %s.", s.Id.Hex()), e
		}
		if e = db.UpdateTime(s); e != nil {
			return fmt.Sprintf("Could not evaluate submission %s.", s.Id.Hex()), e
		}
	}
	return "Successfully evaluated submissions.", nil
}

//Login signs a user into the web app.
func Login(r *http.Request, c *Context) (string, error) {
	un, p, m, e := getCredentials(r)
	if e != nil {
		return m, e
	}
	u, e := db.User(un)
	if e != nil {
		return fmt.Sprintf("User %s not found.", un), e
	}
	if !util.Validate(u.Password, u.Salt, p) {
		e = fmt.Errorf("Invalid username or password.")
		return e.Error(), e
	}
	c.AddUser(un)
	return "Logged in successfully.", nil
}

//Register registers a new user with Impendulo.
func Register(r *http.Request, c *Context) (string, error) {
	un, p, m, e := getCredentials(r)
	if e != nil {
		return m, e
	}
	if e = db.Add(db.USERS, user.New(un, p)); e != nil {
		return fmt.Sprintf("User %s already exists.", un), e
	}
	c.AddUser(un)
	return "Registered successfully.", nil
}

//DeleteUser removes a user and all data associated with them from the system.
func DeleteUser(r *http.Request, c *Context) (string, error) {
	un, m, e := getString(r, "username")
	if e != nil {
		return m, e
	}
	if e = db.RemoveUserById(un); e != nil {
		return fmt.Sprintf("Could not delete user %s.", un), e
	}
	return fmt.Sprintf("Successfully deleted user %s.", un), nil
}

//Logout logs a user out of the system.
func Logout(r *http.Request, c *Context) (string, error) {
	c.RemoveUser()
	return "Successfully logged out.", nil
}

//EditUser
func EditUser(r *http.Request, c *Context) (string, error) {
	id, e := GetString(r, "user-id")
	if e != nil {
		return "Could not read old username.", e
	}
	n, e := GetString(r, "user-name")
	if e != nil {
		return "Could not read new username.", e
	}
	a, e := GetInt(r, "user-perm")
	if e != nil {
		return "Could not read user access level.", e
	}
	if !user.ValidPermission(a) {
		e = fmt.Errorf("invalid user access level %d", a)
		return e.Error(), e
	}
	if id != n {
		if e = db.RenameUser(id, n); e != nil {
			return fmt.Sprintf("could not rename user %s to %s.", id, n), e
		}
	}
	if e = db.Update(db.USERS, bson.M{db.ID: n}, bson.M{db.SET: bson.M{user.ACCESS: a}}); e != nil {
		return "Could not edit user.", e
	}
	return "Successfully edited user.", nil

}

//ImportData
func ImportData(r *http.Request, c *Context) (string, error) {
	n, e := GetString(r, "db")
	if e != nil {
		return "Could not read db to import to.", e
	}
	_, d, e := ReadFormFile(r, "data")
	if e != nil {
		return "Unable to read data file.", e
	}
	if e = mongo.ImportData(n, d); e != nil {
		return "Unable to import db data.", e
	}
	return "Successfully imported db data.", nil
}
