//Contains processing functions used by handlers.go to add or retrieve data.
package webserver

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func RunTool(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	tool, err := GetString(req, "tool")
	if err != nil {
		return
	}
	submissions, err := db.GetSubmissions(bson.M{project.PROJECT_ID: projectId}, bson.M{project.ID: 1})
	if err != nil {
		return
	}
	var runAll bool
	if req.FormValue("runempty-check") == "true" {
		runAll = false
	} else {
		runAll = true
	}
	for _, submission := range submissions {
		files, err := db.GetFiles(bson.M{project.SUBID: submission.Id}, bson.M{project.DATA: 0})
		if err != nil {
			util.Log(err)
			continue
		}
		err = processing.StartSubmission(submission.Id)
		if err != nil {
			util.Log(err)
			continue
		}
		for _, file := range files {
			if resultId, ok := file.Results[tool]; ok && runAll {
				err = db.RemoveResultById(resultId)
				if err != nil {
					util.Log(resultId, err)
					continue
				}
				delete(file.Results, tool)
				change := bson.M{db.SET: bson.M{project.RESULTS: file.Results}}
				err = db.Update(db.FILES, bson.M{project.ID: file.Id}, change)
				if err != nil {
					util.Log(err)
					continue
				}
			}
			err = processing.AddFile(file)
			if err != nil {
				util.Log(err)
			}
		}
		err = processing.EndSubmission(submission.Id)
		if err != nil {
			util.Log(err)
		}
	}
	return
}

//SubmitArchive adds an Intlola archive to the database.
func SubmitArchive(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	uname, err := GetString(req, "user")
	if err != nil {
		return
	}
	if !db.Contains(db.USERS, bson.M{user.ID: uname}) {
		err = fmt.Errorf("User %q not found.", uname)
		return
	}
	_, archiveBytes, err := ReadFormFile(req, "archive")
	if err != nil {
		return
	}
	sub := project.NewSubmission(projectId, uname, project.ARCHIVE_MODE,
		util.CurMilis())
	err = db.AddSubmission(sub)
	if err != nil {
		return
	}
	file := project.NewArchive(sub.Id, archiveBytes)
	err = db.AddFile(file)
	if err != nil {
		return
	}
	//Send file to be analysed.
	err = processing.StartSubmission(sub.Id)
	if err != nil {
		return
	}
	err = processing.AddFile(file)
	if err != nil {
		return
	}
	err = processing.EndSubmission(sub.Id)
	return
}

//ChangeSkeleton replaces a project's skeleton file.
func ChangeSkeleton(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	_, data, err := ReadFormFile(req, "skeleton")
	if err != nil {
		return
	}
	err = db.Update(db.PROJECTS, bson.M{project.ID: projectId},
		bson.M{db.SET: bson.M{project.SKELETON: data}})
	return
}

//AddProject creates a new Impendulo Project.
func AddProject(req *http.Request, ctx *Context) (err error) {
	name, err := GetString(req, "name")
	if err != nil {
		return
	}
	lang, err := GetString(req, "lang")
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	_, skeletonBytes, err := ReadFormFile(req, "skeleton")
	if err != nil {
		return
	}
	p := project.NewProject(name, username, lang, skeletonBytes)
	err = db.AddProject(p)
	return
}

//ReadFormFile reads a file's name and data from a request form.
func ReadFormFile(req *http.Request, name string) (fname string, data []byte, err error) {
	file, header, err := req.FormFile(name)
	if err != nil {
		return
	}
	fname = header.Filename
	data, err = ioutil.ReadAll(file)
	return
}

//Login signs a user into the web app.
func Login(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "username")
	if err != nil {
		return
	}
	pword, err := GetString(req, "password")
	if err != nil {
		return
	}
	u, err := db.GetUserById(uname)
	if err != nil {
		return
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("Invalid username or password.")
	} else {
		ctx.AddUser(uname)
	}
	return
}

//Register registers a new user with Impendulo.
func Register(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "username")
	if err != nil {
		return
	}
	pword, err := GetString(req, "password")
	if err != nil {
		return
	}
	u := user.New(uname, pword)
	err = db.AddUser(u)
	if err == nil {
		ctx.AddUser(uname)
	}
	return
}

//DeleteProject removes a project and all data associated with it from the system.
func DeleteProject(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err == nil {
		err = db.RemoveProjectById(projectId)
	}
	return
}

//DeleteUser removes a user and all data associated with them from the system.
func DeleteUser(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "user")
	if err == nil {
		err = db.RemoveUserById(uname)
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
	if tipe == "project" {
		var pid bson.ObjectId
		pid, err = util.ReadId(idStr)
		if err != nil {
			return
		}
		ctx.Browse.Pid = pid
		ctx.Browse.IsUser = false
		subs, err = db.GetSubmissions(
			bson.M{project.PROJECT_ID: pid}, nil, "-"+project.TIME)
	} else if tipe == "user" {
		ctx.Browse.Uid = idStr
		ctx.Browse.IsUser = true
		subs, err = db.GetSubmissions(
			bson.M{project.USER: ctx.Browse.Uid}, nil, "-"+project.TIME)
	} else {
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

//GetInt retrieves an integer value from a request form.
func GetInt(req *http.Request, name string) (found int, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.Atoi(iStr)
	return
}

func GetLines(req *http.Request, name string) []int {
	start, err := GetInt(req, name+"focusstart")
	if err != nil {
		err = nil
		start = 0
	}
	end, err := GetInt(req, name+"focusend")
	if err != nil {
		err = nil
		end = start
	}
	lines := make([]int, end-start+1)
	for i := start; i <= end; i++ {
		lines[i-start] = i
	}
	return lines
}

//GetStrings retrieves a string value from a request form.
func GetStrings(req *http.Request, name string) (vals []string, err error) {
	if req.Form == nil {
		err = req.ParseForm()
		if err != nil {
			return
		}
	}
	vals = req.Form[name]
	return
}

//GetString retrieves a string value from a request form.
func GetString(req *http.Request, name string) (val string, err error) {
	val = req.FormValue(name)
	if strings.TrimSpace(val) == "" {
		err = fmt.Errorf("Invalid value for %s.", name)
	}
	return
}

//GetResultData retrieves a DisplayResult for a given file and result name.
func GetResultData(resultName string, fileId bson.ObjectId) (res tool.DisplayResult, err error) {
	var file *project.File
	matcher := bson.M{project.ID: fileId}
	file, err = db.GetFile(matcher, nil)
	if err != nil {
		return
	}
	switch resultName {
	case tool.CODE:
		res = tool.NewCodeResult(file.Data)
	case diff.NAME:
		res = diff.NewDiffResult(file)
	case tool.SUMMARY:
		res = tool.NewSummaryResult()
		//Load summary for each available result.
		for name, resid := range file.Results {
			var currentRes tool.ToolResult
			currentRes, err = db.GetToolResult(name,
				bson.M{project.ID: resid}, nil)
			if err != nil {
				return
			}
			res.(*tool.SummaryResult).AddSummary(currentRes)
		}
	default:
		ival, ok := file.Results[resultName]
		if !ok {
			res = tool.NewErrorResult(
				fmt.Errorf("No result available for %v.", resultName))
			return
		}
		switch val := ival.(type) {
		case bson.ObjectId:
			//Retrieve result from the db.
			matcher = bson.M{project.ID: val}
			res, err = db.GetDisplayResult(resultName,
				matcher, nil)
		case string:
			//Error, so create new error result.
			switch val {
			case tool.TIMEOUT:
				res = new(tool.TimeoutResult)
			case tool.NORESULT:
				res = new(tool.NoResult)
			default:
				res = tool.NewErrorResult(
					fmt.Errorf("No result available for %v.", resultName))
			}
		default:
			res = tool.NewErrorResult(
				fmt.Errorf("No result available for %v.", resultName))
		}
	}
	return
}

func getFile(id bson.ObjectId) (file *project.File, err error) {
	selector := bson.M{project.NAME: 1, project.TIME: 1}
	file, err = db.GetFile(bson.M{project.ID: id}, selector)
	return
}

func getIndex(req *http.Request, name string, maxSize int) (ret int, err error) {
	ret, err = GetInt(req, name)
	if err != nil {
		return
	}
	if ret > maxSize {
		ret = 0
	} else if ret < 0 {
		ret = maxSize
	}
	return
}

func getSelected(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "currentIndex", maxSize)
}

func getNeighbour(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "nextIndex", maxSize)
}

func projectName(id bson.ObjectId) (name string, err error) {
	proj, err := db.GetProject(bson.M{project.ID: id},
		bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	name = proj.Name
	return
}
