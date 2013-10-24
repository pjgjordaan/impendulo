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
	"errors"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

//SubmitArchive adds an Intlola archive to the database.
func SubmitArchive(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	username, msg, err := getUser(ctx)
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
	username, msg, err := getUser(ctx)
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
	u, err := GetString(req, "user")
	if err != nil {
		msg = "Could not read user."
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
	subId, err := util.ReadId(req.FormValue("subid"))
	if err != nil {
		msg = "Could not read submission id."
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
	u, err := GetString(req, "user")
	if err != nil {
		msg = "Could not read user."
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
	fileId, err := util.ReadId(req.FormValue("fileid"))
	if err != nil {
		msg = "Could not read file id."
		return
	}
	subId, err := util.ReadId(req.FormValue("subid"))
	if err != nil {
		msg = "Could not read submission id."
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

//RetrieveFileInfo fetches all filenames in a submission.
func RetrieveFileInfo(req *http.Request, ctx *Context) (ret []*db.FileInfo, err error) {
	err = ctx.Browse.SetSid(req)
	if err != nil {
		return
	}
	matcher := bson.M{db.SUBID: ctx.Browse.Sid, db.TYPE: project.SRC}
	ret, err = db.FileInfos(matcher)
	if err != nil {
		return
	}
	sub, err := db.Submission(
		bson.M{db.ID: ctx.Browse.Sid},
		bson.M{db.PROJECTID: 1, db.USER: 1},
	)
	if err != nil {
		return
	}
	ctx.Browse.Pid = sub.ProjectId
	ctx.Browse.Uid = sub.User
	return
}

//Snapshots retrieves snapshots of a given file in a submission.
func Snapshots(subId bson.ObjectId, fileName string) (ret []*project.File, err error) {
	matcher := bson.M{db.SUBID: subId,
		db.TYPE: project.SRC, db.NAME: fileName}
	selector := bson.M{db.TIME: 1, db.SUBID: 1}
	ret, err = db.Files(matcher, selector, db.TIME)
	if err == nil && len(ret) == 0 {
		err = fmt.Errorf("No files found with name %q.", fileName)
	}
	return
}

//RetrieveSubmissions fetches all submissions in a project or by a user.
func RetrieveSubmissions(req *http.Request, ctx *Context) ([]*project.Submission, error) {
	perr := ctx.Browse.SetPid(req)
	uerr := ctx.Browse.SetUid(req)
	if perr == nil {
		ctx.Browse.IsUser = false
		return db.Submissions(
			bson.M{db.PROJECTID: ctx.Browse.Pid},
			nil, "-"+db.TIME,
		)
	} else if uerr == nil {
		ctx.Browse.IsUser = true
		return db.Submissions(
			bson.M{db.USER: ctx.Browse.Uid},
			nil, "-"+db.TIME,
		)
	} else {
		return nil, errors.New("No id found.")
	}
}

//LoadSkeleton makes a project skeleton available for download.
func LoadSkeleton(req *http.Request) (path string, err error) {
	idStr := req.FormValue("project")
	projectId, err := util.ReadId(idStr)
	if err != nil {
		return
	}
	name := strconv.FormatInt(time.Now().Unix(), 10)
	path = filepath.Join(util.BaseDir(), "skeletons", idStr, name+".zip")
	//If the skeleton is saved for downloading we don't need to store it again.
	if util.Exists(path) {
		return
	}
	p, err := db.Project(bson.M{db.ID: projectId}, nil)
	if err != nil {
		return
	}
	//Save file to filesystem and return path to it.
	err = util.SaveFile(path, p.Skeleton)
	return
}

//getFile
func getFile(id bson.ObjectId) (file *project.File, err error) {
	selector := bson.M{db.NAME: 1, db.TIME: 1}
	file, err = db.File(bson.M{db.ID: id}, selector)
	return
}

//projectName retrieves the project name associated with the project identified by id.
func projectName(id bson.ObjectId) (name string, err error) {
	proj, err := db.Project(bson.M{db.ID: id},
		bson.M{db.NAME: 1})
	if err != nil {
		return
	}
	name = proj.Name
	return
}

func EvaluateSubmissions(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	all := req.FormValue("project") == "all"
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
