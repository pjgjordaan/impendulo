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
	sub := project.NewSubmission(projectId, username, project.ARCHIVE_MODE, util.CurMilis())
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
	jpf := project.NewJPFFile(projectId, jpfHeader.Filename, username, jpfBytes)
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

func retrieveNames(req *http.Request, ctx *context.Context) (ret []string, msg string, err error) {
	ctx.Browse.Sid = req.FormValue("subid")
	if !bson.IsObjectIdHex(ctx.Browse.Sid) {
		err = fmt.Errorf("Invalid submission id %q", ctx.Browse.Sid)
		msg = err.Error()
		return
	}
	subId := bson.ObjectIdHex(ctx.Browse.Sid)
	matcher := bson.M{project.SUBID: subId, project.TYPE: project.SRC}
	ret, err = db.GetFileNames(matcher)
	if err != nil {
		msg = fmt.Sprintf("Could not retrieve filenames for submission.")
	}
	return
}

func retrieveFiles(req *http.Request, ctx *context.Context) (ret []*project.File, msg string, err error) {
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
	selector := bson.M{project.NAME: 1, project.ID: 1, project.RESULTS: 1, project.TIME: 1}
	file, err = db.GetFile(bson.M{project.ID: id}, selector)
	if err != nil {
		msg = fmt.Sprintf("Could not retrieve file.")
	}
	return
}

func getSelected(req *http.Request, maxSize int) (selected int, msg string, err error) {
	selStr := req.FormValue("selected")
	selected, err = strconv.Atoi(selStr)
	if err != nil {
		msg = fmt.Sprintf("Invalid index %q.", selStr)
		return
	}
	if selected >= maxSize {
		err = fmt.Errorf("Index size %q too big.", selected)
		msg = err.Error()
	}
	return
}

func getResult(req *http.Request, fileId bson.ObjectId) (res tool.Result, msg string, err error) {
	name := req.FormValue("resultname")
	var file *project.File
	if strings.ToLower(name) == "code" {
		file, err = db.GetFile(bson.M{project.ID: fileId}, bson.M{project.DATA: 1})
		if err != nil {
			msg = fmt.Sprintf("Could not retrieve code.")
			return
		}
		res = tool.NewCodeResult(fileId, file.Data)
	} else {
		file, err = db.GetFile(bson.M{project.ID: fileId}, bson.M{project.RESULTS: 1})
		if err != nil {
			msg = fmt.Sprintf("Could not retrieve file results.")
			return
		}
		id, ok := file.Results[name]
		if !ok {
			res = tool.NewErrorResult(fmt.Sprintf("Could not retrieve result for %q.", name))

		}
		selector := bson.M{project.DATA: 1}
		matcher := bson.M{project.ID: id}
		if strings.HasPrefix(javac.NAME, name) {
			res, err = db.GetJavacResult(matcher, selector)
		} else if strings.HasPrefix(junit.NAME, name) {
			res, err = db.GetJUnitResult(matcher, selector)
		} else if strings.HasPrefix(jpf.NAME, name) {
			res, err = db.GetJPFResult(matcher, selector)
		} else if strings.HasPrefix(findbugs.NAME, name) {
			res, err = db.GetFindbugsResult(matcher, selector)
		} else if strings.HasPrefix(pmd.NAME, name) {
			res, err = db.GetPMDResult(matcher, selector)
		}else if strings.HasPrefix(checkstyle.NAME, name) {
			res, err = db.GetCheckstyleResult(matcher, selector)
		} else {
			err = fmt.Errorf("Unknown result %q.", name)
		}
		if err != nil {
			msg = fmt.Sprintf("Could not retrieve result.")
		}
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

func projectName(idStr string) string {
	if !bson.IsObjectIdHex(idStr) {
		return ""
	}
	id := bson.ObjectIdHex(idStr)
	proj, err := db.GetProject(bson.M{project.ID: id}, bson.M{project.NAME: 1})
	if err != nil {
		return ""
	}
	return proj.Name
}
