//Contains processing functions used by handlers.go to add or retrieve data.
package webserver

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/jpf"
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

//A function used to add data to the database.
type PostFunc func(*http.Request, *Context) error

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
	processing.StartSubmission(sub.Id)
	processing.AddFile(file)
	processing.EndSubmission(sub.Id)
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

//AddTest adds a new test to a project.
func AddTest(req *http.Request, ctx *Context) (err error) {
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
		//Read data files if provided.
		_, dataBytes, err = ReadFormFile(req, "data")
		if err != nil {
			return
		}
	}
	//Read package name from file.
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	username, err := ctx.Username()
	if err != nil {
		return
	}
	test := project.NewTest(projectId, testName, username,
		pkg, testBytes, dataBytes)
	err = db.AddTest(test)
	return
}

//AddJPF replaces a project's JPF configuration file.
func AddJPF(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	_, data, err := ReadFormFile(req, "jpf")
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	jpfConfig := jpf.NewConfig(projectId, username, data)
	err = db.AddJPF(jpfConfig)
	return
}

//CreateJPF replaces a project's JPF configuration file.
func CreateJPF(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	listeners, err := GetStrings(req, "addedL")
	if err != nil {
		return
	}
	search, err := GetString(req, "addedS")
	if err != nil {
		return
	}
	vals := map[string][]string{
		"search.class": []string{search},
		"listener":     listeners,
	}
	other, err := GetString(req, "other")
	if err == nil {
		props := readProperties(other)
		for k, v := range props {
			vals[k] = v
		}
	}
	data, err := jpf.JPFBytes(vals)
	if err != nil {
		return
	}
	jpfConfig := jpf.NewConfig(projectId, username, data)
	err = db.AddJPF(jpfConfig)
	return
}

func readProperties(raw string) (props map[string][]string) {
	props = make(map[string][]string)
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		params := strings.Split(util.RemoveEmpty(line), "=")
		if len(params) == 2 {
			key, val := params[0], params[1]
			if len(key) > 0 && len(val) > 0 && jpf.Allowed(key) {
				split := strings.Split(val, ",")
				vals := make([]string, 0, len(split))
				for _, v := range split {
					if v != "" {
						vals = append(vals, v)
					}
				}
				if v, ok := props[key]; ok {
					props[key] = append(v, vals...)
				} else {
					props[key] = vals
				}
			}
		}
	}
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

//RetrieveNames fetches all filenames in a submission.
func RetrieveNames(req *http.Request, ctx *Context) (ret []string, err error) {
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
		//Load project id if browsing in user view.
		var sub *project.Submission
		sub, err = db.GetSubmission(bson.M{project.ID: subId},
			bson.M{project.PROJECT_ID: 1})
		if err != nil {
			return
		}
		ctx.Browse.Pid = sub.ProjectId.Hex()
	}
	return
}

//RetrieveFileInfo fetches all filenames in a submission.
func RetrieveFileInfo(req *http.Request, ctx *Context) (ret []*db.FileInfo, err error) {
	ctx.Browse.Sid = req.FormValue("subid")
	subId, err := util.ReadId(ctx.Browse.Sid)
	if err != nil {
		return
	}
	matcher := bson.M{project.SUBID: subId, project.TYPE: project.SRC}
	ret, err = db.GetFileInfo(matcher)
	if err != nil {
		return
	}
	sub, err := db.GetSubmission(bson.M{project.ID: subId},
		bson.M{project.PROJECT_ID: 1, project.USER: 1})
	if err != nil {
		return
	}
	ctx.Browse.Pid = sub.ProjectId.Hex()
	ctx.Browse.Uid = sub.User
	return
}

//RetrieveFiles fetches all files in a submission with a given name.
func RetrieveFiles(req *http.Request, ctx *Context) (ret []*project.File, err error) {
	name, err := GetString(req, "filename")
	if err != nil {
		return
	}
	sid, err := util.ReadId(ctx.Browse.Sid)
	if err != nil {
		return
	}
	matcher := bson.M{project.SUBID: sid,
		project.TYPE: project.SRC, project.NAME: name}
	selector := bson.M{project.TIME: 1}
	ret, err = db.GetFiles(matcher, selector, project.TIME)
	if err == nil && len(ret) == 0 {
		err = fmt.Errorf("No files found with name %q.", name)
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
		ctx.Browse.Pid = idStr
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
	if err != nil {
		return
	}
	return
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
	var fileSelector bson.M
	matcher := bson.M{project.ID: fileId}
	if resultName == tool.CODE {
		//Need to load file's source code (data).
		fileSelector = bson.M{project.DATA: 1}
	} else {
		fileSelector = bson.M{project.RESULTS: 1}
	}
	file, err = db.GetFile(matcher, fileSelector)
	if err != nil {
		return
	}
	if resultName == tool.CODE {
		res = tool.NewCodeResult(file.Data)
	} else if resultName == tool.SUMMARY {
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
	} else {
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
				matcher, bson.M{project.DATA: 1})
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

func projectName(idStr string) (name string, err error) {
	var id bson.ObjectId
	id, err = util.ReadId(idStr)
	if err != nil {
		return
	}
	var proj *project.Project
	proj, err = db.GetProject(bson.M{project.ID: id},
		bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	name = proj.Name
	return
}
