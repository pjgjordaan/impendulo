package web

import (
	"labix.org/v2/mgo/bson"
	"net/http"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/context"
	"fmt"
	"io/ioutil"
	"strings"
)

func getTabs(ctx *context.Context)string{
	if ctx.Session.Values["user"] != nil{
		return "adminTabs.html"
	} 
	return "outTabs.html"
}

func processArchive(req *http.Request, ctx *context.Context) error{
	proj := req.FormValue("project")
	if !bson.IsObjectIdHex(proj) {
		return fmt.Errorf("Error parsing selected project %q", proj)
	} 
	projectId := bson.ObjectIdHex(proj)
	archiveFile, archiveHeader, err :=  req.FormFile("archive")
	if err != nil{
		return fmt.Errorf("Error loading archive file")
	}
	archiveBytes, err := ioutil.ReadAll(archiveFile)
	if err != nil{
		return fmt.Errorf("Error reading archive file %q.", archiveHeader.Filename)
	}
	username, err := ctx.Username()
	if err != nil{
		return err
	}
	fmt.Println(projectId, archiveBytes, username)
	return nil
}


func processTest(req *http.Request, ctx *context.Context) error{
	proj := req.FormValue("project")
	if !bson.IsObjectIdHex(proj) {
		return fmt.Errorf("Error parsing selected project %q", proj)
	} 
	projectId := bson.ObjectIdHex(proj)
	testFile, testHeader, err :=  req.FormFile("test")
	if err != nil{
		return fmt.Errorf("Error loading test file")
	}
	dataFile, dataHeader,err :=  req.FormFile("data")
	if err != nil{
		return fmt.Errorf("Error loading data files")
	}
	testBytes, err := ioutil.ReadAll(testFile)
	if err != nil{
		return fmt.Errorf("Error reading test file %q.", testHeader.Filename)
	}
	dataBytes, err := ioutil.ReadAll(dataFile)
	if err != nil{
		return fmt.Errorf("Error reading data files %q.", dataHeader.Filename)
	}
	pkg := util.GetPackage(testFile)
	username, err := ctx.Username()
	if err != nil{
		return err
	}
	test := project.NewTest(projectId, testHeader.Filename, username, pkg, testBytes, dataBytes)
	return db.AddTest(test)
}

func processProject(req *http.Request, ctx *context.Context)error{
	name, lang := strings.TrimSpace(req.FormValue("name")), strings.TrimSpace(req.FormValue("lang"))
	if name == ""{
		return fmt.Errorf("Invalid project name")
	}
	if lang == ""{
		return fmt.Errorf("Invalid language")
	}
	username, err := ctx.Username()
	if err != nil{
		return err
	}
	p := project.NewProject(name, username, lang)
	err = db.AddProject(p)
	if err != nil {
		return fmt.Errorf("Error adding project %q.", name)
	}
	err = ctx.LoadProjects()
	if err != nil {
		return fmt.Errorf("Could not load projects.")
	}
	return nil
}

type DisplayResult struct{
	Name string
	Code string
	Results []*tool.Result
}

func buildResults(req *http.Request)(*DisplayResult, error){
	fileId := req.FormValue("fileid")
	if !bson.IsObjectIdHex(fileId){
		return nil, fmt.Errorf("Could not retrieve file.")
	}
	f, err := db.GetFile(bson.M{project.ID: bson.ObjectIdHex(fileId)}, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve file.")
	}
	res := &DisplayResult{Name: f.InfoStr(project.NAME), Results: make([]*tool.Result, len(f.Results))}
	if f.Type() == project.SRC{
		res.Code = strings.TrimSpace(string(f.Data))
	}
	i := 0
	for k,v := range f.Results{
		res.Results[i], err = db.GetResult(bson.M{project.ID: v}, nil)
		if err != nil{
			return res, fmt.Errorf("No result found for %q.", k)
		}
		i++
	}
	return res, nil
}

func processLogin(req *http.Request)(string, error){
	uname, pword := strings.TrimSpace(req.FormValue("username")), strings.TrimSpace(req.FormValue("password")) 
	u, err := db.GetUserById(uname)
	if err != nil{
		return uname, fmt.Errorf("User %q is not registered.", uname)
	}else if !util.Validate(u.Password, u.Salt, pword){
		return uname, fmt.Errorf("Invalid username or password.")
	}
	return u.Name, nil
}

func processRegistration(req *http.Request)(string, error){
	uname, pword := strings.TrimSpace(req.FormValue("username")), strings.TrimSpace(req.FormValue("password")) 
	if uname == ""{
		return uname, fmt.Errorf("Invalid username.")
	}
	if pword == ""{
		return uname, fmt.Errorf("Invalid password.")
	}
	u := user.NewUser(uname, pword)
	err := db.AddUser(u)
	if err != nil{
		return uname, fmt.Errorf("User %q already exists.", uname)
	}
	return uname, nil
}
