package web

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/context"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
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

func doArchive(req *http.Request, ctx *context.Context) (string, error) {
	proj := req.FormValue("project")
	if !bson.IsObjectIdHex(proj) {
		err := fmt.Errorf("Error parsing selected project %q.", proj)
		return err.Error(), err
	}
	projectId := bson.ObjectIdHex(proj)
	archiveFile, archiveHeader, err := req.FormFile("archive")
	if err != nil {
		return fmt.Sprintf("Error loading archive file."), err
	}
	archiveBytes, err := ioutil.ReadAll(archiveFile)
	if err != nil {
		return fmt.Sprintf("Error reading archive file %q.", archiveHeader.Filename), err
	}
	username, err := ctx.Username()
	if err != nil {
		return err.Error(), err
	}
	sub := project.NewSubmission(projectId, username, project.ARCHIVE_MODE)
	err = db.AddSubmission(sub)
	if err != nil {
		return fmt.Sprintf("Could not create submission."), err
	}
	file := project.NewArchive(sub.Id, archiveBytes, project.ZIP)
	err = db.AddFile(file)
	if err != nil {
		return fmt.Sprintf("Could not create submission."), err
	}
	processing.StartSubmission(sub)
	processing.AddFile(file)
	processing.EndSubmission(sub)
	return fmt.Sprintf("Submission successful."), nil
}

func doTest(req *http.Request, ctx *context.Context) (string, error) {
	proj := req.FormValue("project")
	if !bson.IsObjectIdHex(proj) {
		err := fmt.Errorf("Error parsing selected project %q", proj)
		return err.Error(), err
	}
	projectId := bson.ObjectIdHex(proj)
	testFile, testHeader, err := req.FormFile("test")
	if err != nil {
		return fmt.Sprintf("Error loading test file"), err
	}
	testBytes, err := ioutil.ReadAll(testFile)
	if err != nil {
		return fmt.Sprintf("Error reading test file %q.", testHeader.Filename), err
	}
	hasData := req.FormValue("data-check")
	var dataBytes []byte
	if hasData == "" {
		dataBytes = make([]byte, 0)
	} else if hasData == "true" {
		dataFile, dataHeader, err := req.FormFile("data")
		if err != nil {
			return fmt.Sprintf("Error loading data files."), err
		}
		dataBytes, err = ioutil.ReadAll(dataFile)
		if err != nil {
			return fmt.Sprintf("Error reading data files %q.", dataHeader.Filename), err
		}
	}
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	username, err := ctx.Username()
	if err != nil {
		return err.Error(), err
	}
	test := project.NewTest(projectId, testHeader.Filename, username, pkg, testBytes, dataBytes)
	err = db.AddTest(test)
	if err != nil {
		return fmt.Sprintf("Unable to add test %q.", testHeader.Filename), err
	}
	return fmt.Sprintf("Successfully added test %q.", testHeader.Filename), err
}

func doJPF(req *http.Request, ctx *context.Context) (string, error) {
	proj := req.FormValue("project")
	if !bson.IsObjectIdHex(proj) {
		err := fmt.Errorf("Error parsing selected project %q.", proj)
		return err.Error(), err
	}
	projectId := bson.ObjectIdHex(proj)
	jpfFile, jpfHeader, err := req.FormFile("jpf")
	if err != nil {
		return fmt.Sprintf("Error loading jpf config file."), err
	}
	jpfBytes, err := ioutil.ReadAll(jpfFile)
	if err != nil {
		return fmt.Sprintf("Error reading jpf config file %q.", jpfHeader.Filename), err
	}
	username, err := ctx.Username()
	if err != nil {
		return err.Error(), err
	}
	jpf := project.NewJPF(projectId, jpfHeader.Filename, username, jpfBytes)
	if jpf.IsJava{
		jpf.Package = util.GetPackage(bytes.NewReader(jpf.Data))
	}
	err = db.AddJPF(jpf)
	if err != nil {
		return fmt.Sprintf("Unable to add jpf config file %q.", jpf.Name), err
	}
	return fmt.Sprintf("Successfully added jpf config file %q.", jpf.Name), err
}

func doProject(req *http.Request, ctx *context.Context) (string, error) {
	name, lang := strings.TrimSpace(req.FormValue("name")), strings.TrimSpace(req.FormValue("lang"))
	if name == "" {
		err := fmt.Errorf("Invalid project name.")
		return err.Error(), err
	}
	if lang == "" {
		err := fmt.Errorf("Invalid language.")
		return err.Error(), err
	}
	username, err := ctx.Username()
	if err != nil {
		return err.Error(), err
	}
	p := project.NewProject(name, username, lang)
	err = db.AddProject(p)
	if err != nil {
		return fmt.Sprintf("Error adding project %q.", name), err
	}
	return "Successfully added project.", nil
}

func doLogin(req *http.Request, ctx *context.Context) (string, error) {
	uname, pword := strings.TrimSpace(req.FormValue("username")), strings.TrimSpace(req.FormValue("password"))
	u, err := db.GetUserById(uname)
	if err != nil {
		return fmt.Sprintf("User %q is not registered.", uname), err
	} else if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("Invalid username or password.")
		return err.Error(), err
	}
	ctx.AddUser(uname)
	return fmt.Sprintf("Successfully logged in as %q.", uname), nil
}

func doRegister(req *http.Request, ctx *context.Context) (string, error) {
	uname, pword := strings.TrimSpace(req.FormValue("username")), strings.TrimSpace(req.FormValue("password"))
	if uname == "" {
		err := fmt.Errorf("Invalid username.")
		return err.Error(), err
	}
	if pword == "" {
		err := fmt.Errorf("Invalid password.")
		return err.Error(), err
	}
	u := user.NewUser(uname, pword)
	err := db.AddUser(u)
	if err != nil {
		return fmt.Sprintf("User %q already exists.", uname), err
	}
	ctx.AddUser(uname)
	return fmt.Sprintf("Successfully registered as %q.", uname), nil
}

func retrieveFiles(req *http.Request, ctx *context.Context) ([]*project.File, string, error) {
	ctx.Browse.Sid = req.FormValue("subid")
	if !bson.IsObjectIdHex(ctx.Browse.Sid) {
		err := fmt.Errorf("Invalid submission id %q", ctx.Browse.Sid)
		return nil, err.Error(), err
	}
	subId := bson.ObjectIdHex(ctx.Browse.Sid)
	var err error
	choices := []bson.M{bson.M{project.INFO + "." + project.TYPE: project.EXEC}, bson.M{project.INFO + "." + project.TYPE: project.SRC}}
	files, err := db.GetFiles(bson.M{project.SUBID: subId, "$or": choices}, bson.M{project.INFO: 1})
	if err != nil {
		return nil, fmt.Sprintf("Could not retrieve files for submission."), err
	}
	return files, "", nil
}

type DisplayResult struct {
	Name    string
	Code    string
	Results bson.M
}

func buildResults(req *http.Request) (*DisplayResult, string, error) {
	fileId := req.FormValue("fileid")
	if !bson.IsObjectIdHex(fileId) {
		err := fmt.Errorf("Could not retrieve file.")
		return nil, err.Error(), err
	}
	file, err := db.GetFile(bson.M{project.ID: bson.ObjectIdHex(fileId)}, nil)
	if err != nil {
		return nil, fmt.Sprintf("Could not retrieve file."), err
	}
	res := &DisplayResult{Name: file.Name, Results: file.Results}
	if file.Type == project.SRC {
		res.Code = strings.TrimSpace(string(file.Data))
	}
	return res, "Successfully retrieved results.", nil
}

func getResult(id interface{})(*tool.Result, error){
	return db.GetResult(bson.M{project.ID: id}, nil)	
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
