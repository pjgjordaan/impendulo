package webserver

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
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

func doArchive(req *http.Request, ctx *Context) (err error) {
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
	sub := project.NewSubmission(projectId, uname, project.ARCHIVE_MODE, util.CurMilis())
	err = db.AddSubmission(sub)
	if err != nil {
		return
	}
	file := project.NewArchive(sub.Id, archiveBytes, project.ZIP)
	err = db.AddFile(file)
	if err != nil {
		return
	}
	processing.StartSubmission(sub.Id)
	processing.AddFile(file)	
	processing.EndSubmission(sub.Id)
	return
}

func doSkeleton(req *http.Request, ctx *Context) (err error) {
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

func doTest(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	testName, testBytes, err := ReadFormFile(req, "test")
	if err != nil {
		return
	}
	hasData := req.FormValue("data-check")
	var dataBytes []byte
	if hasData == "" {
		dataBytes = make([]byte, 0)
	} else if hasData == "true" {
		_, dataBytes, err = ReadFormFile(req, "data")
		if err != nil {
			return
		}
	}
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	username, err := ctx.Username()
	if err != nil {
		return
	}
	test := project.NewTest(projectId, testName, username, pkg, testBytes, dataBytes)
	err = db.AddTest(test)
	return
}

func doJPF(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	name, data, err := ReadFormFile(req, "jpf")
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	jpf := project.NewJPFFile(projectId, name, username, data)
	err = db.AddJPF(jpf)
	return
}

func ReadFormFile(req *http.Request, name string)(fname string, data []byte, err error){
	file, header, err := req.FormFile(name)
	if err != nil {
		return
	}
	fname = header.Filename
	data, err = ioutil.ReadAll(file)
	return
}

func doProject(req *http.Request, ctx *Context) (err error) {
	name, err := GetString(req, "name")
	if err != nil {
		return
	}
	lang, err := GetString(req, "lang")
	if err != nil{
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

func doLogin(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "username")
	if err != nil {
		return
	}
	pword, err := GetString(req, "password")
	if err != nil{
		return
	}
	u, err := db.GetUserById(uname)
	if err != nil {
		return
	} 
	if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("Invalid username or password.")
	} else{
		ctx.AddUser(uname)
	}
	return
}

func doRegister(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "username")
	if err != nil {
		return
	}
	pword, err := GetString(req, "password")
	if err != nil{
		return
	}
	u := user.NewUser(uname, pword)
	err = db.AddUser(u)
	if err == nil {
		ctx.AddUser(uname)
	}
	return
}

func doDeleteProject(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err == nil {
		err = db.RemoveProjectById(projectId)
	}
	return
}

func doDeleteUser(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "user")
	if err == nil {
		err = db.RemoveUserById(uname)
	}
	return
}

func retrieveNames(req *http.Request, ctx *Context) (ret []string, err error) {
	ctx.Browse.Sid = req.FormValue("subid")
	subId, err := util.ReadId(ctx.Browse.Sid)
	if err != nil {
		return
	}
	matcher := bson.M{project.SUBID: subId, project.TYPE: project.SRC}
	ret, err = db.GetFileNames(matcher)
	if err != nil {
		return
	}
	if ctx.Browse.IsUser {
		var sub *project.Submission
		sub, err = db.GetSubmission(bson.M{project.ID: subId}, bson.M{project.PROJECT_ID: 1})
		if err != nil {
			return
		} 
		ctx.Browse.Pid = sub.ProjectId.Hex()
	}
	return
}

func getCompileData(files []*project.File) (ret []bool) {
	ret = make([]bool, len(files))
	for i, file := range files {
		file, _ = db.GetFile(bson.M{project.ID: file.Id}, nil)
		if id, ok := file.Results[javac.NAME]; ok {
			res, _ := db.GetJavacResult(bson.M{project.ID: id}, nil)
			ret[i] = res.Success()
		} else {
			ret[i] = false
		}
	}
	return ret
}

func retrieveFiles(req *http.Request, ctx *Context) (ret []*project.File, err error) {
	name, err := GetString(req, "filename")
	if err != nil{
		return
	}
	sid, err := util.ReadId(ctx.Browse.Sid)
	if err != nil{
		return
	}
	matcher := bson.M{project.SUBID: sid, project.TYPE: project.SRC, project.NAME: name}
	selector := bson.M{project.ID: 1, project.NAME: 1}
	ret, err = db.GetFiles(matcher, selector, project.NUM)
	if err == nil && len(ret) == 0 {
		err = fmt.Errorf("No files found with name %q.", name)
	}
	return
}

func getFile(id bson.ObjectId) (file *project.File, err error) {
	selector := bson.M{project.NAME: 1, project.ID: 1, project.TIME: 1, project.NUM: 1}
	file, err = db.GetFile(bson.M{project.ID: id}, selector)
	return
}

func getSelected(req *http.Request, maxSize int) (int, error) {
	return GetInt(req, "currentIndex", maxSize)
}

func getNeighbour(req *http.Request, maxSize int) (int, bool) {
	val, err := GetInt(req, "nextIndex", maxSize)
	return val, err == nil
}

func retrieveSubmissions(req *http.Request, ctx *Context) (subs []*project.Submission, err error) {
	tipe, err := GetString(req, "type")
	if err != nil{
		return
	}
	idStr, err := GetString(req, "id")
	if err != nil{
		return
	}
	if tipe == "project" {
		var pid bson.ObjectId
		pid, err = util.ReadId(idStr)
		if err != nil {
			return
		}
		ctx.Browse.Pid = idStr
		ctx.Browse.IsUser = false
		subs, err = db.GetSubmissions(bson.M{project.PROJECT_ID: pid}, nil)
	} else if tipe == "user" {
		ctx.Browse.Uid = idStr
		ctx.Browse.IsUser = true
		subs, err = db.GetSubmissions(bson.M{project.USER: ctx.Browse.Uid}, nil)
	} else{
		err = fmt.Errorf("Unknown request type %q", tipe)
	}
	return
}

func projectName(idStr string) (name string, err error) {
	var id bson.ObjectId
	id, err = util.ReadId(idStr)
	if err != nil {
		return
	}
	var proj *project.Project
	proj, err = db.GetProject(bson.M{project.ID: id}, bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	name = proj.Name
	return
}

func loadSkeleton(req *http.Request) (path string, err error) {
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
	err = util.SaveFile(path, p.Skeleton)
	return
}

func GetInt(req *http.Request, name string, maxSize int) (found int, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.Atoi(iStr)
	if err != nil {
		return
	}
	if found > maxSize {
		err = fmt.Errorf("Integer size %v too big.", found)
	}
	return
}

func GetString(req *http.Request, name string) (val string, err error) {
	val = req.FormValue(name)
	if strings.TrimSpace(val) == "" {
		err = fmt.Errorf("Invalid value for %s.", name)
	}
	return
}

func GetResultData(resultName string, fileId bson.ObjectId) (res tool.DisplayResult, err error) {
	var file *project.File
	dataSelector := bson.M{project.DATA: 1}
	matcher := bson.M{project.ID: fileId}
	if resultName == tool.CODE {
		file, err = db.GetFile(matcher, dataSelector)
		if err != nil {
			return
		}
		res = tool.NewCodeResult(fileId, file.Data)
	} else if resultName == tool.SUMMARY {
		file, err = db.GetFile(matcher, bson.M{project.RESULTS: 1})
		if err != nil {
			return
		}
		res = tool.NewSummaryResult()
		for name, resid := range file.Results{
			var currentRes tool.ToolResult
			currentRes, err = db.GetToolResult(name, bson.M{project.ID: resid},nil) 
			if err != nil {
				return
			}
			res.(*tool.SummaryResult).AddSummary(currentRes)
		}
	} else {
		file, err = db.GetFile(matcher, bson.M{project.RESULTS: 1})
		if err != nil {
			return
		}
		id, ok := file.Results[resultName]
		if !ok {
			res = tool.NewErrorResult(fileId, fmt.Errorf("No result available for %v.", resultName))
			return
		}
		matcher = bson.M{project.ID: id}
		res, err = db.GetDisplayResult(resultName, matcher, dataSelector)
	}
	return
}
