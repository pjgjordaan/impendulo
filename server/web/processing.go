package web

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/context"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/tool/checkstyle"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"mime/multipart"
)

func getNav(ctx *context.Context) string {
	if _, err := ctx.Username(); err != nil {
		return "outNav.html"
	}
	return "inNav.html"
}

type processor func(*http.Request, *context.Context) (string, error)

func (p processor) exec(req *http.Request, ctx *context.Context) error {
	msg, err := p(req, ctx)
	ctx.AddMessage(msg, err != nil)
	return err
}

func doArchive(req *http.Request, ctx *context.Context) (msg string, err error) {
	projectId, err := ReadId(req.FormValue("project")) 
	if err != nil {
		msg = err.Error()
		return
	}
	archiveFile, archiveHeader, err := req.FormFile("archive")
	if err != nil {
		msg = fmt.Sprintf("Error loading archive file.")
		return
	}
	archiveBytes, err := ioutil.ReadAll(archiveFile)
	if err != nil {
		msg = fmt.Sprintf("Error reading archive file %q.", archiveHeader.Filename)
		return
	}
	username, err := ctx.Username()
	if err != nil {
		msg = err.Error()
		return
	}
	sub := project.NewSubmission(projectId, username, project.ARCHIVE_MODE, util.CurMilis())
	err = db.AddSubmission(sub)
	if err != nil {
		msg = fmt.Sprintf("Could not create submission.")
		return
	}
	file := project.NewArchive(sub.Id, archiveBytes, project.ZIP)
	err = db.AddFile(file)
	if err != nil {
		msg = fmt.Sprintf("Could not create file.")
		return
	}
	processing.StartSubmission(sub)
	processing.AddFile(file)
	processing.EndSubmission(sub)
	msg = fmt.Sprintf("Submission successful.")
	return
}

func doTest(req *http.Request, ctx *context.Context) (msg string, err error) {
	projectId, err := ReadId(req.FormValue("project")) 
	if err != nil {
		msg = err.Error()
		return
	}
	testFile, testHeader, err := req.FormFile("test")
	if err != nil {
		msg = fmt.Sprintf("Error loading test file")
		return
	}
	testBytes, err := ioutil.ReadAll(testFile)
	if err != nil {
		msg = fmt.Sprintf("Error reading test file %q.", testHeader.Filename)
		return
	}
	hasData := req.FormValue("data-check")
	var dataBytes []byte
	if hasData == "" {
		dataBytes = make([]byte, 0)
	} else if hasData == "true" {
		var dataFile multipart.File
		var dataHeader *multipart.FileHeader
		dataFile, dataHeader, err = req.FormFile("data")
		if err != nil {
			msg = fmt.Sprintf("Error loading data files.")
			return
		}
		dataBytes, err = ioutil.ReadAll(dataFile)
		if err != nil {
			msg = fmt.Sprintf("Error reading data files %q.", dataHeader.Filename)
			return
		}
	}
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	username, err := ctx.Username()
	if err != nil {
		msg = err.Error()
		return
	}
	test := project.NewTest(projectId, testHeader.Filename, username, pkg, testBytes, dataBytes)
	err = db.AddTest(test)
	if err != nil {
		msg = fmt.Sprintf("Unable to add test %q.", testHeader.Filename)
		return
	}
	msg = fmt.Sprintf("Successfully added test %q.", testHeader.Filename)
	return
}

func doJPF(req *http.Request, ctx *context.Context) (msg string, err error) {
	projectId, err := ReadId(req.FormValue("project")) 
	if err != nil {
		msg = err.Error()
		return
	}
	jpfFile, jpfHeader, err := req.FormFile("jpf")
	if err != nil {
		msg = fmt.Sprintf("Error loading jpf config file.")
		return
	}
	jpfBytes, err := ioutil.ReadAll(jpfFile)
	if err != nil {
		msg = fmt.Sprintf("Error reading jpf config file %q.", jpfHeader.Filename)
		return
	}
	username, err := ctx.Username()
	if err != nil {
		msg = err.Error()
		return
	}
	jpf := project.NewJPFFile(projectId, jpfHeader.Filename, username, jpfBytes)
	err = db.AddJPF(jpf)
	if err != nil {
		msg = fmt.Sprintf("Unable to add jpf config file %q.", jpf.Name)
		return
	}
	msg = fmt.Sprintf("Successfully added jpf config file %q.", jpf.Name)
	return
}

func doProject(req *http.Request, ctx *context.Context) (msg string, err error) {
	name, lang := strings.TrimSpace(req.FormValue("name")), strings.TrimSpace(req.FormValue("lang"))
	if name == "" {
		err = fmt.Errorf("Invalid project name.")
		msg = err.Error()
		return
	}
	if lang == "" {
		err = fmt.Errorf("Invalid language.")
		msg = err.Error()
		return
	}
	username, err := ctx.Username()
	if err != nil {
		msg = err.Error()
		return
	}
	p := project.NewProject(name, username, lang)
	err = db.AddProject(p)
	if err != nil {
		msg = fmt.Sprintf("Error adding project %q.", name)
		return
	}
	msg = "Successfully added project."
	return
}

func doLogin(req *http.Request, ctx *context.Context) (msg string, err error) {
	uname, pword := strings.TrimSpace(req.FormValue("username")), strings.TrimSpace(req.FormValue("password"))
	u, err := db.GetUserById(uname)
	if err != nil {
		msg = fmt.Sprintf("User %q is not registered.", uname)
		return
	} else if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("Invalid username or password.")
		msg =  err.Error()
		return
	}
	ctx.AddUser(uname)
	msg = fmt.Sprintf("Successfully logged in as %q.", uname)
	return
}

func doRegister(req *http.Request, ctx *context.Context) (msg string, err error) {
	uname, pword := strings.TrimSpace(req.FormValue("username")), strings.TrimSpace(req.FormValue("password"))
	if uname == "" {
		err = fmt.Errorf("Invalid username.")
		msg = err.Error()
		return
	}
	if pword == "" {
		err = fmt.Errorf("Invalid password.")
		msg = err.Error()
		return
	}
	u := user.NewUser(uname, pword)
	err = db.AddUser(u)
	if err != nil {
		msg = fmt.Sprintf("User %q already exists.", uname)
		return
	}
	ctx.AddUser(uname)
	msg = fmt.Sprintf("Successfully registered as %q.", uname)
	return
}

func retrieveNames(req *http.Request, ctx *context.Context) (ret []string, msg string, err error) {
	ctx.Browse.Sid = req.FormValue("subid")
	subId, err := ReadId(ctx.Browse.Sid)
	if err != nil {
		msg = err.Error()
		return
	}
	matcher := bson.M{project.SUBID: subId, project.TYPE: project.SRC}
	ret, err = db.GetFileNames(matcher)
	if err != nil {
		msg = fmt.Sprintf("Could not retrieve filenames for submission.")
	}
	return
}

func getCompileData(files []*project.File)(ret []bool){
	ret = make([]bool, len(files))
	for i, file := range files{
		file, _ = db.GetFile(bson.M{project.ID: file.Id}, nil)
		if id,ok := file.Results[javac.NAME]; ok{
			res,_ := db.GetJavacResult(bson.M{project.ID: id}, nil)
			ret[i] = res.Success()
		} else{
			ret[i] = false
		}
	}
	return ret
}

func retrieveFiles(req *http.Request, ctx *context.Context) (ret []*project.File, msg string, err error) {
	req.ParseForm()
	fmt.Println(req.Form)
	name := req.FormValue("filename")
	if !bson.IsObjectIdHex(ctx.Browse.Sid) {
		err = fmt.Errorf("Invalid submission id %q.", ctx.Browse.Sid)
		msg = err.Error()
		return
	}
	matcher := bson.M{project.SUBID: bson.ObjectIdHex(ctx.Browse.Sid), project.TYPE: project.SRC, project.NAME: name}
	selector := bson.M{project.ID: 1, project.NAME: 1}
	ret, err = db.GetFiles(matcher, selector, project.NUM)
	if err != nil {
		msg = fmt.Sprintf("Could not retrieve files for submission.")
	}
	if len(ret) == 0 {
		err = fmt.Errorf("No files found with name %q.", name)
		msg = err.Error()
	}
	return
}

func getFile(id bson.ObjectId) (file *project.File, msg string, err error) {
	selector := bson.M{project.NAME: 1, project.ID: 1, project.RESULTS: 1, project.TIME: 1, project.NUM:1}
	file, err = db.GetFile(bson.M{project.ID: id}, selector)
	if err != nil {
		msg = fmt.Sprintf("Could not retrieve file.")
	}
	return
}

func getSelected(req *http.Request, maxSize int) (int, string, error) {
	return GetInt(req, "currentIndex", maxSize)
}

func getNeighbour(req *http.Request, maxSize int) (int, string, error) {
	return GetInt(req, "nextIndex", maxSize)
}


func getResult(req *http.Request, fileId bson.ObjectId) (res tool.Result, msg string, err error) {
	name := req.FormValue("resultname")
	res, err = GetResultData(name, fileId)
	if err != nil{
		msg = fmt.Sprintf("Could not retrieve result %q.", name)
	}
	return
}

func retrieveSubmissions(req *http.Request, ctx *context.Context) (subs []*project.Submission, msg string, err error) {
	tipe := req.FormValue("type")
	idStr := req.FormValue("id")
	if tipe == "project" {
		if !bson.IsObjectIdHex(idStr) {
			err = fmt.Errorf("Invalid id %q", idStr)
			msg = err.Error()
			return
		}
		ctx.Browse.Pid = idStr
		ctx.Browse.IsUser = false
		pid := bson.ObjectIdHex(idStr)
		subs, err = db.GetSubmissions(bson.M{project.PROJECT_ID: pid}, nil)
		if err != nil {
			msg = "Could not retrieve project submissions."
		}
		return
	} else if tipe == "user" {
		ctx.Browse.Uid = idStr
		ctx.Browse.IsUser = true
		subs, err = db.GetSubmissions(bson.M{project.USER: ctx.Browse.Uid}, nil)
		if err != nil {
			msg = "Could not retrieve user submissions."
		}
		return
	}
	err = fmt.Errorf("Unknown request type %q", tipe)
	msg = err.Error()
	return
}

func projectName(idStr string) (name string,err error) {
	var id bson.ObjectId
	id, err = ReadId(idStr)
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

func ReadId(idStr string)(ret bson.ObjectId, err error){
	if !bson.IsObjectIdHex(idStr) {
		err = fmt.Errorf("Invalid id string %q.", idStr)
		return
	}
	ret = bson.ObjectIdHex(idStr)
	return
}

func GetInt(req *http.Request, name string, maxSize int) (found int, msg string, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.Atoi(iStr)
	if err != nil {
		msg = fmt.Sprintf("Invalid int %q.", iStr)
		return
	}
	if found > maxSize {
		err = fmt.Errorf("Integer size %q too big.", found)
		msg = err.Error()
	}
	return
}

func GetResultData(resultName string, fileId bson.ObjectId) (res tool.Result, err error) {
	var file *project.File
	selector := bson.M{project.DATA: 1}
	matcher := bson.M{project.ID: fileId}
	if strings.ToLower(resultName) == "code" {
		file, err = db.GetFile(matcher, selector)
		if err != nil {
			return
		}
		res = tool.NewCodeResult(fileId, file.Data)
	} else {
		file, err = db.GetFile(matcher, bson.M{project.RESULTS: 1})
		if err != nil {
			return
		}
		id, ok := file.Results[resultName]
		if !ok {
			res = tool.NewErrorResult(fmt.Errorf("Could not retrieve result for %q.", resultName))
			return
		}
		matcher = bson.M{project.ID: id}
		if strings.HasPrefix(javac.NAME, resultName) {
			res, err = db.GetJavacResult(matcher, selector)
		} else if strings.HasPrefix(junit.NAME, resultName) {
			res, err = db.GetJUnitResult(matcher, selector)
		} else if strings.HasPrefix(jpf.NAME, resultName) {
			res, err = db.GetJPFResult(matcher, selector)
		} else if strings.HasPrefix(findbugs.NAME, resultName) {
			res, err = db.GetFindbugsResult(matcher, selector)
		} else if strings.HasPrefix(pmd.NAME, resultName) {
			res, err = db.GetPMDResult(matcher, selector)
		}else if strings.HasPrefix(checkstyle.NAME, resultName) {
			res, err = db.GetCheckstyleResult(matcher, selector)
		} else {
			err = fmt.Errorf("Unknown result %q.", resultName)
		}
	}
	return
}