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
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/mongo"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	//A function used to fullfill a POST request.
	Poster func(*http.Request, *context.C) (string, error)
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
		"addproject": AddProject, "addskeleton": AddSkeleton, "addarchive": AddArchive,
		"runtools": RunTools, "deleteprojects": DeleteProjects, "deleteusers": DeleteUsers, "deleteassignments": DeleteAssignments,
		"deletesubmissions": DeleteSubmissions, "deleteresults": DeleteResults, "deleteskeletons": DeleteSkeletons,
		"deletetests": DeleteTests, "importdata": ImportData, "renamefiles": RenameFiles, "login": Login, "register": Register,
		"logout": Logout, "editproject": EditProject, "edituser": EditUser, "editsubmission": EditSubmission,
		"editfile": EditFile, "edittest": EditTest, "addassignment": AddAssignment, "editassignment": EditAssignment,
		"deletecache": ClearCache,
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
		router.Add("POST", "/"+n, Handler(f.CreatePost(indexPosts[n]))).Name(n)
	}
}

//CreatePost loads a post request handler.
func (p Poster) CreatePost(index bool) Handler {
	return func(w http.ResponseWriter, r *http.Request, c *context.C) error {
		m, e := p(r, c)
		c.AddMessage(m, e != nil)
		if e == nil && index {
			http.Redirect(w, r, getRoute("index"), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		}
		return e
	}
}

func AddAssignment(r *http.Request, c *context.C) (string, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return "Could not read project id.", e
	}
	sid, e := webutil.Id(r, "skeleton-id")
	if e != nil {
		return "Could not read skeleton id.", e
	}
	n, e := webutil.String(r, "assignment-name")
	if e != nil {
		return "Could not read assignment name.", e
	}
	un, e := c.Username()
	if e != nil {
		return "Could not retrieve user.", e
	}
	s, e := webutil.Int64(r, "assignment-start")
	if e != nil {
		return "Could not read assignment start.", e
	}
	end, e := webutil.Int64(r, "assignment-end")
	if e != nil {
		return "Could not read assignment end.", e
	}
	if e = db.Add(db.ASSIGNMENTS, project.NewAssignment(pid, sid, n, un, s, end)); e != nil {
		return "Could not add assignment.", e
	}
	return "Successfully added assignment.", nil
}

func AddSkeleton(r *http.Request, c *context.C) (string, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return "Could not read project id.", e
	}
	n, e := webutil.String(r, "skeletonname")
	if e != nil {
		return "Could not read skeleton name.", e
	}
	_, s, e := webutil.FileData(r, "skeleton")
	if e != nil {
		return "Could not read skeleton file.", e
	}
	if e = db.Add(db.SKELETONS, project.NewSkeleton(pid, n, s)); e != nil {
		return "Could not add skeleton.", e
	}
	return "Successfully added skeleton.", nil
}

func AddArchive(r *http.Request, c *context.C) (string, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return "Could not read project id.", e
	}
	aid, e := webutil.Id(r, "assignment-id")
	if e != nil {
		return "Could not read assignment id.", e
	}
	u, e := webutil.String(r, "user-id")
	if e != nil {
		return "Could not read user.", e
	}
	_, a, e := webutil.FileData(r, "archive")
	if e != nil {
		return "Could not read archive.", e
	}
	//We need to create a submission for this archive so that
	//it can be added to the db and so that it can be processed
	s := project.NewSubmission(pid, aid, u, project.ARCHIVE_MODE, util.CurMilis())
	if e = db.Add(db.SUBMISSIONS, s); e != nil {
		return "Could not create submission.", e
	}
	f := project.NewArchive(s.Id, a)
	if e = db.Add(db.FILES, f); e != nil {
		return "Could not store archive.", e
	}
	if e := mq.StartSubmission(s.Id); e != nil {
		return "Could not start archive submission.", e
	}
	if e = mq.AddFile(f); e != nil {
		return "Could not submit archive.", e
	}
	if e = mq.EndSubmission(s.Id); e != nil {
		return "Could not complete archive submission.", e
	}
	return "Archive submitted successfully.", nil
}

//AddProject creates a new Impendulo Project.
func AddProject(r *http.Request, c *context.C) (string, error) {
	n, e := webutil.String(r, "projectname")
	if e != nil {
		return "Could not read project name.", e
	}
	l, e := webutil.Language(r, "lang")
	if e != nil {
		return "Could not read project language.", e
	}
	un, e := c.Username()
	if e != nil {
		return "Could not retrieve user.", e
	}
	d, e := webutil.String(r, "description")
	if e != nil {
		return "Could not read description.", e
	}
	if e = db.Add(db.PROJECTS, project.New(n, un, d, l)); e != nil {
		return "Could not add project.", e
	}
	return "Successfully added project.", nil
}

//DeleteProjects removes a project and all data associated with it from the system.
func DeleteProjects(r *http.Request, c *context.C) (string, error) {
	pids, e := webutil.Strings(r, "project-id")
	if e != nil {
		return "Could not read projects.", e
	}
	for _, p := range pids {
		id, e := convert.Id(p)
		if e != nil {
			util.Log(e)
			continue
		}
		if e = db.RemoveProjectById(id); e != nil {
			util.Log(e)
		}
	}
	return "Successfully deleted project.", nil
}

//DeleteUsers removes users and all data associated with them from the system.
func DeleteUsers(r *http.Request, c *context.C) (string, error) {
	us, e := webutil.Strings(r, "user-id")
	if e != nil {
		return "Could not read user.", e
	}
	for _, u := range us {
		if e = db.RemoveUserById(u); e != nil {
			util.Log(e)
		}
	}
	return "Successfully deleted users.", nil
}

func DeleteSubmissions(r *http.Request, c *context.C) (string, error) {
	ss, e := webutil.Strings(r, "submission-id")
	if e != nil {
		return "Could not read submissions.", e
	}
	for _, s := range ss {
		id, e := convert.Id(s)
		if e != nil {
			util.Log(e)
			continue
		}
		if e = db.RemoveSubmissionById(id); e != nil {
			util.Log(e)
		}
	}
	return "Successfully deleted submissions.", nil
}

func DeleteAssignments(r *http.Request, c *context.C) (string, error) {
	as, e := webutil.Strings(r, "assignment-id")
	if e != nil {
		return "Could not read assignment.", e
	}
	for _, a := range as {
		id, e := convert.Id(a)
		if e != nil {
			util.Log(e)
			continue
		}
		if e = db.RemoveAssignmentById(id); e != nil {
			util.Log(e)
		}
	}
	return "Successfully deleted assignments.", nil
}

func DeleteSkeletons(r *http.Request, c *context.C) (string, error) {
	sks, e := webutil.Strings(r, "skeleton-id")
	if e != nil {
		return "Could not read skeletons.", e
	}
	for _, sk := range sks {
		id, e := convert.Id(sk)
		if e != nil {
			util.Log(e)
			continue
		}
		if e = db.RemoveById(db.SKELETONS, id); e != nil {
			util.Log(e)
		}
	}
	return "Successfully deleted skeletons.", nil
}

func DeleteTests(r *http.Request, c *context.C) (string, error) {
	ts, e := webutil.Strings(r, "test-id")
	if e != nil {
		return "Could not read tests.", e
	}
	for _, t := range ts {
		id, e := convert.Id(t)
		if e != nil {
			util.Log(e)
			continue
		}
		if e = db.RemoveById(db.TESTS, id); e != nil {
			util.Log(e)
		}
	}
	return "Successfully deleted tests.", nil
}

//DeleteResults removes all results for a specic project.
func DeleteResults(r *http.Request, c *context.C) (string, error) {
	ss, e := webutil.Strings(r, "submission-id")
	if e != nil {
		return "Could not read submissions.", e
	}
	for _, s := range ss {
		sid, e := convert.Id(s)
		if e != nil {
			util.Log(e)
			continue
		}
		fs, e := db.Files(bson.M{db.SUBID: sid}, bson.M{db.DATA: 0}, 0)
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
func EditProject(r *http.Request, c *context.C) (string, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return "Could not read project id.", e
	}
	p, e := db.Project(bson.M{db.ID: pid}, nil)
	if e != nil {
		return "Could not load project.", e
	}
	sm := bson.M{}
	if n, e := webutil.String(r, "project-name"); e == nil && n != p.Name {
		sm[db.NAME] = n
	}
	if u, e := webutil.String(r, "project-user"); e == nil && p.User != u && db.Contains(db.USERS, bson.M{db.ID: u}) {
		sm[db.USER] = u
	}
	if l, e := webutil.Language(r, "project-lang"); e == nil && project.Supported(l) && p.Lang != l {
		sm[db.LANG] = l
	}
	if d, e := webutil.String(r, "project-description"); e == nil && p.Description != d {
		sm[db.DESCRIPTION] = d
	}
	if len(sm) == 0 {
		return "Nothing to update", nil
	}
	if e = db.Update(db.PROJECTS, bson.M{db.ID: pid}, bson.M{db.SET: sm}); e != nil {
		return "Could not edit project.", e
	}
	return "Successfully edited project.", nil
}

func EditAssignment(r *http.Request, c *context.C) (string, error) {
	aid, e := webutil.Id(r, "assignment-id")
	if e != nil {
		return "Could not read assignment id.", e
	}
	a, e := db.Assignment(bson.M{db.ID: aid}, nil)
	if e != nil {
		return "Could not load assignment.", e
	}
	m := bson.M{}
	if pid, e := webutil.Id(r, "assignment-project"); e == nil && a.ProjectId != pid && db.Contains(db.PROJECTS, bson.M{db.ID: pid}) {
		m[db.PROJECTID] = pid
	}
	if sid, e := webutil.Id(r, "assignment-skeleton"); e == nil && a.SkeletonId != sid && db.Contains(db.SKELETONS, bson.M{db.ID: sid}) {
		m[db.SKELETONID] = sid
	}
	if u, e := webutil.String(r, "assignment-user"); e == nil && a.User != u && db.Contains(db.USERS, bson.M{db.ID: u}) {
		m[db.USER] = u
	}
	if n, e := webutil.String(r, "assignment-name"); e == nil && a.Name != n {
		m[db.NAME] = n
	}
	if len(m) == 0 {
		return "Nothing to update", nil
	}
	if e = db.Update(db.ASSIGNMENTS, bson.M{db.ID: aid}, bson.M{db.SET: m}); e != nil {
		return "Could not edit assignment.", e
	}
	return "Successfully edited assignment.", nil
}

//EditSubmission
func EditSubmission(r *http.Request, c *context.C) (string, error) {
	sid, e := webutil.Id(r, "submission-id")
	if e != nil {
		return "Could not read submission id.", e
	}
	s, e := db.Submission(bson.M{db.ID: sid}, nil)
	if e != nil {
		return "Could not load submission.", e
	}
	sm := bson.M{}
	if pid, e := webutil.Id(r, "submission-project"); e == nil && s.ProjectId != pid && db.Contains(db.PROJECTS, bson.M{db.ID: pid}) {
		sm[db.PROJECTID] = pid
	}
	if aid, e := webutil.Id(r, "submission-assignment"); e == nil && s.AssignmentId != aid {
		sm[db.ASSIGNMENTID] = aid
	}
	if u, e := webutil.String(r, "submission-user"); e == nil && s.User != u && db.Contains(db.USERS, bson.M{db.ID: u}) {
		sm[db.USER] = u
	}
	if len(sm) == 0 {
		return "Nothing to update", nil
	}
	if e = db.Update(db.SUBMISSIONS, bson.M{db.ID: sid}, bson.M{db.SET: sm}); e != nil {
		return "Could not edit submission.", e
	}
	return "Successfully edited submission.", nil
}

//EditFile
func EditFile(r *http.Request, c *context.C) (string, error) {
	fid, e := webutil.Id(r, "file-id")
	if e != nil {
		return "Could not read file id.", e
	}
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return "Could not load file.", e
	}
	sm := bson.M{}
	if n, e := webutil.String(r, "file-name"); e == nil && f.Name != n {
		sm[db.NAME] = n
	}
	if p, e := webutil.String(r, "file-package"); e == nil && f.Package != p {
		sm[db.PKG] = p
	}
	if len(sm) == 0 {
		return "Nothing to update", nil
	}
	if e = db.Update(db.FILES, bson.M{db.ID: fid}, bson.M{db.SET: sm}); e != nil {
		return "Could not edit file.", e
	}
	return "Successfully edited file.", nil
}

func EditTest(r *http.Request, c *context.C) (string, error) {
	tid, e := webutil.Id(r, "test-id")
	if e != nil {
		return "Could not read test id.", e
	}
	t, e := db.JUnitTest(bson.M{db.ID: tid}, nil)
	if e != nil {
		return "Could not load test.", e
	}
	sm := bson.M{}
	if pid, e := webutil.Id(r, "test-project"); e == nil && pid != t.ProjectId && db.Contains(db.PROJECTS, bson.M{db.ID: pid}) {
		sm[db.PROJECTID] = pid
	}
	if n, e := webutil.String(r, "test-name"); e == nil && t.Name != n {
		sm[db.NAME] = n
	}
	if p, e := webutil.String(r, "test-package"); e == nil && t.Package != p {
		sm[db.PKG] = p
	}
	if tn, e := webutil.String(r, "test-target-name"); e == nil {
		tp, _ := webutil.String(r, "test-target-package")
		if t.Target == nil || (tp != t.Target.Package || tn != t.Target.FullName()) {
			sm[db.TARGET] = tool.NewTarget(tn, tp, "", project.JAVA)
		}
	}
	if len(sm) == 0 {
		return "Nothing to update", nil
	}
	if e = db.Update(db.TESTS, bson.M{db.ID: tid}, bson.M{db.SET: sm}); e != nil {
		return "Could not edit file.", e
	}
	return "Successfully edited test.", nil
}

//Login signs a user into the web app.
func Login(r *http.Request, c *context.C) (string, error) {
	un, p, e := webutil.Credentials(r)
	if e != nil {
		return "Could not retrieve credentials.", e
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
func Register(r *http.Request, c *context.C) (string, error) {
	un, p, e := webutil.Credentials(r)
	if e != nil {
		return "Could not retrieve credentials.", e
	}
	if e = db.Add(db.USERS, user.New(un, p)); e != nil {
		return fmt.Sprintf("User %s already exists.", un), e
	}
	c.AddUser(un)
	return "Registered successfully.", nil
}

//Logout logs a user out of the system.
func Logout(r *http.Request, c *context.C) (string, error) {
	c.RemoveUser()
	return "Successfully logged out.", nil
}

//EditUser
func EditUser(r *http.Request, c *context.C) (string, error) {
	id, e := webutil.String(r, "user-id")
	if e != nil {
		return "Could not read old username.", e
	}
	u, e := db.User(id)
	if e != nil {
		return "Could not load user.", e
	}
	n, e := webutil.String(r, "user-name")
	if e != nil {
		return "Could not read new username.", e
	}
	a, e := webutil.Int(r, "user-perm")
	if e != nil {
		return "Could not read user access level.", e
	}
	if !user.ValidPermission(a) {
		e = fmt.Errorf("invalid user access level %d", a)
		return e.Error(), e
	}
	p := user.Permission(a)
	if id == n && u.Access == p {
		return "Nothing to update.", nil
	}
	if id != n {
		if e = db.RenameUser(id, n); e != nil {
			return fmt.Sprintf("could not rename user %s to %s.", id, n), e
		}
	}
	if u.Access != p {
		if e = db.Update(db.USERS, bson.M{db.ID: n}, bson.M{db.SET: bson.M{user.ACCESS: a}}); e != nil {
			return "Could not edit user.", e
		}
	}
	return "Successfully edited user.", nil

}

//ImportData
func ImportData(r *http.Request, c *context.C) (string, error) {
	n, e := webutil.String(r, "db")
	if e != nil {
		return "Could not read db to import to.", e
	}
	_, d, e := webutil.FileData(r, "data")
	if e != nil {
		return "Unable to read data file.", e
	}
	if e = mongo.ImportData(n, d); e != nil {
		return "Unable to import db data.", e
	}
	return "Successfully imported db data.", nil
}

func RenameFiles(r *http.Request, c *context.C) (string, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return "Could not read project.", e
	}
	nn, e := webutil.String(r, "file-name-new")
	if e != nil {
		return "Could not read new file name.", e
	}
	np, _ := webutil.String(r, "package-name-new")
	on, e := webutil.String(r, "file-name-old")
	if e != nil {
		return "Could not read current file name.", e
	}
	op, _ := webutil.String(r, "package-name-old")
	subs, e := db.Submissions(bson.M{db.PROJECTID: pid}, nil)
	if e != nil {
		return "Could not retrieve submissions.", e
	}
	for _, s := range subs {
		fs, e := db.Files(bson.M{db.SUBID: s.Id, db.NAME: on, db.PKG: op}, bson.M{db.ID: 1}, 0)
		if e != nil {
			return "Could not retrieve files.", e
		}
		for _, fi := range fs {
			f, e := db.File(bson.M{db.ID: fi.Id}, nil)
			if e != nil {
				return "Could not retrieve file.", e
			}
			f.Rename(np, nn)
			if e := db.Update(db.FILES, bson.M{db.ID: f.Id}, bson.M{db.SET: bson.M{db.DATA: f.Data, db.NAME: f.Name, db.PKG: f.Package}}); e != nil {
				return "Could not update file name.", e
			}
		}
	}
	return fmt.Sprintf("Succesfully renamed files to package: %s class: %s.", np, nn), nil
}

func ClearCache(r *http.Request, c *context.C) (string, error) {
	if e := db.RemoveCollection(db.CALC); e != nil {
		return "Could not clear cache.", e
	}
	return "Cache cleared.", nil
}
