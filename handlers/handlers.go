package handlers

import (
	"labix.org/v2/mgo/bson"
	"net/http"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/util"
	"github.com/godfried/cabanga/project"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/impendulo/context"
	"fmt"
	"io/ioutil"
	"strings"
)

func homeView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T("homeView.html", "noRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true})
}

func projectView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	langs := []string{"Java"}
	return T("projectView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "p":true, "langs": langs})
}

func testView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	if ctx.Projects == nil{
		err := ctx.LoadProjects()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("testView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "t":true})
}

func archiveView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	if ctx.Projects == nil{
		err := ctx.LoadProjects()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("archiveView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "a":true})
}

func registerView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T("registerView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "r":true})
}

func getUsers(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	if ctx.Users == nil{
		err := ctx.LoadUsers()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("homeView.html", "userRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true})
}

func getProjects(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	if ctx.Projects == nil{
		err := ctx.LoadProjects()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("homeView.html", "projRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true})
}


func getSubmissions(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	tp := req.FormValue("type")
	idStr := req.FormValue("id")
	subs, err := ctx.Subs(tp, idStr)
	if err != nil {
		ctx.AddError(err.Error())
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return nil
	}
	return T("homeView.html", "subRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true, "subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	subId := req.FormValue("subid")
	fileRes, err := ctx.GetFiles(subId)
	if err != nil {
		ctx.AddError(err.Error())
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return nil
	}
	return T("homeView.html", "fileRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true, "fileRes": fileRes})
}

func getResults(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	res, err := buildResults(req)
	if err != nil {
		ctx.AddError(err.Error())
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return nil
	}
	return T("homeView.html", "dispRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true, "res": res})
}

func login(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	uname, err := processLogin(req)
	if err != nil{
		ctx.AddError(err.Error())
	}else{
		ctx.Session.Values["user"] = uname
	}
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func register(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	uname, err := processRegistration(req)
	if err != nil{
		ctx.AddError(err.Error())
		http.Redirect(w, req, reverse("registerview"), http.StatusSeeOther)
	} else{
		ctx.AddSuccess(fmt.Sprintf("Successfully registered as %q.", uname))
		ctx.Session.Values["user"] = uname
		http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	}
	return nil
}

func logout(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func addTest(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processTest(req, ctx)
	if err != nil{
		ctx.AddError(err.Error())
	}else{
		ctx.AddSuccess("Successfully added test.")
	}
	http.Redirect(w, req, reverse("testview"), http.StatusSeeOther)
	return nil
}

func addProject(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processProject(req, ctx)
	if err != nil{
		ctx.AddError(err.Error())
	} else{
		ctx.AddSuccess("Successfully added project.")
	}
	http.Redirect(w, req, reverse("projectview"), http.StatusSeeOther)
	return nil
}

func submitArchive(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processArchive(req, ctx)
	if err != nil{
		ctx.AddError(err.Error())
	}else{
		ctx.AddSuccess("Successfully submitted archive.")
	}
	http.Redirect(w, req, reverse("archiveview"), http.StatusSeeOther)
	return nil
}

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
	test := project.NewTest(projectId, testHeader.Filename, username, pkg, testBytes, dataBytes)
	return db.AddTest(test)
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
	pkg := getPackage(testFile)
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
