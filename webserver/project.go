package webserver

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
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
	if !db.Contains(db.USERS, bson.M{user.ID: username}) {
		err = fmt.Errorf("User %s not found.", username)
		msg = err.Error()
		return
	}
	_, archiveBytes, err := ReadFormFile(req, "archive")
	if err != nil {
		msg = "Could not read archive."
		return
	}
	sub := project.NewSubmission(projectId, username, project.ARCHIVE_MODE,
		util.CurMilis())
	err = db.AddSubmission(sub)
	if err != nil {
		msg = "Could not create submission."
		return
	}
	file := project.NewArchive(sub.Id, archiveBytes)
	err = db.AddFile(file)
	if err != nil {
		msg = "Could not store archive."
		return
	}
	//Send file to be analysed.
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
	err = db.Update(db.PROJECTS, bson.M{project.ID: projectId},
		bson.M{db.SET: bson.M{project.SKELETON: data}})
	if err != nil {
		msg = "Could not update skeleton file."
	} else {
		msg = "Successfully updated skeleton file."
	}
	return
}

//AddProject creates a new Impendulo Project.
func AddProject(req *http.Request, ctx *Context) (msg string, err error) {
	name, err := GetString(req, "name")
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
	err = db.AddProject(p)
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

//RetrieveFileInfo fetches all filenames in a submission.
func RetrieveFileInfo(req *http.Request, ctx *Context) (ret []*db.FileInfo, err error) {
	subId, serr := util.ReadId(req.FormValue("subid"))
	if serr == nil {
		ctx.Browse.Sid = subId
	}
	matcher := bson.M{project.SUBID: ctx.Browse.Sid, project.TYPE: project.SRC}
	ret, err = db.GetFileInfo(matcher)
	if err != nil {
		return
	}
	sub, err := db.GetSubmission(bson.M{project.ID: ctx.Browse.Sid},
		bson.M{project.PROJECT_ID: 1, project.USER: 1})
	if err != nil {
		return
	}
	ctx.Browse.Pid = sub.ProjectId
	ctx.Browse.Uid = sub.User
	return
}

//RetrieveFiles fetches all files in a submission with a given name.
func RetrieveFiles(req *http.Request, ctx *Context) (ret []*project.File, err error) {
	name, ferr := GetString(req, "filename")
	if ferr == nil {
		ctx.Browse.FileName = name
	}
	matcher := bson.M{project.SUBID: ctx.Browse.Sid,
		project.TYPE: project.SRC, project.NAME: ctx.Browse.FileName}
	selector := bson.M{project.TIME: 1}
	ret, err = db.GetFiles(matcher, selector, project.TIME)
	if err == nil && len(ret) == 0 {
		err = fmt.Errorf("No files found with name %q.", ctx.Browse.FileName)
	}
	return
}

//RetrieveSubmissions fetches all submissions in a project or by a user.
func RetrieveSubmissions(req *http.Request, ctx *Context) (subs []*project.Submission, err error) {
	tipe, err := GetString(req, "type")
	if err != nil {
		return
	}
	idStr, err := GetString(req, "id")
	if err != nil {
		return
	}
	switch tipe {
	case "project":
		var pid bson.ObjectId
		pid, err = util.ReadId(idStr)
		if err != nil {
			return
		}
		ctx.Browse.Pid = pid
		ctx.Browse.IsUser = false
		subs, err = db.GetSubmissions(
			bson.M{project.PROJECT_ID: pid}, nil, "-"+project.TIME)
	case "user":
		ctx.Browse.Uid = idStr
		ctx.Browse.IsUser = true
		subs, err = db.GetSubmissions(
			bson.M{project.USER: ctx.Browse.Uid}, nil, "-"+project.TIME)
	default:
		err = fmt.Errorf("Unknown request type %q", tipe)
	}
	return
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
	if util.Exists(path) {
		return
	}
	p, err := db.GetProject(bson.M{project.ID: projectId}, nil)
	if err != nil {
		return
	}
	//Save file to filesystem and return path to it.
	err = util.SaveFile(path, p.Skeleton)
	return
}

//getFile
func getFile(id bson.ObjectId) (file *project.File, err error) {
	selector := bson.M{project.NAME: 1, project.TIME: 1}
	file, err = db.GetFile(bson.M{project.ID: id}, selector)
	return
}

//projectName
func projectName(id bson.ObjectId) (name string, err error) {
	proj, err := db.GetProject(bson.M{project.ID: id},
		bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	name = proj.Name
	return
}
