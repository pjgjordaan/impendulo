package webserver

import (
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"net/http"
	"net/url"
	"strings"
)

var store sessions.Store

const LOG_HANDLERS = "webserver/handlers.go"

func init() {
	store = sessions.NewCookieStore(util.CookieKeys())
}

func getNav(ctx *Context) string {
	if _, err := ctx.Username(); err != nil {
		return "outNavbar"
	}
	return "inNavbar"
}

//Handler is used to handle incoming requests.
//It allows for better session management.
type Handler func(http.ResponseWriter, *http.Request, *Context) error

//ServeHTTP loads a the current session, handles  the request and
//then stores the session.
func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		util.Log(err, LOG_HANDLERS)
	}
	ctx := NewContext(sess)
	buf := new(HttpBuffer)
	err = checkAccess(req.URL, ctx)
	if err != nil {
		ctx.AddMessage(err.Error(), true)
		err = nil
		http.Redirect(buf, req, getRoute("index"), http.StatusSeeOther)
	} else {
		err = h(buf, req, ctx)
	}
	if err != nil {
		util.Log(err, LOG_HANDLERS)
	}
	if err = ctx.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.Apply(w)
}

func checkAccess(url *url.URL, ctx *Context) error {
	start := strings.LastIndex(url.Path, "/") + 1
	end := strings.Index(url.Path, "?")
	if end < 0 {
		end = len(url.Path)
	}
	if start > end {
		return fmt.Errorf("Invalid request %s", url.Path)
	}
	name := url.Path[start:end]
	val, ok := perms[name]
	if !ok {
		return fmt.Errorf("Could not find request %s", url.Path)
	}
	if val == 0 {
		return nil
	}
	if !ctx.LoggedIn() {
		return fmt.Errorf("Insufficient permissions to access %s", url.Path)
	}
	return nil
}

var views = map[string]string{
	"homeView": "home", "testView": "submit",
	"skeletonView": "submit", "jpfFileView": "submit",
	"registerView": "register", "projectDownloadView": "download",
	"projectDeleteView": "delete", "userDeleteView": "delete",
	"userResult": "home", "projectResult": "home",
	"jpfConfigView": "submit", "archiveView": "submit",
	"projectView": "submit", "statusView": "status",
}

var perms = map[string]int{
	"homeview": 0, "testview": 1, "skeletonview": 1,
	"jpffileview": 1, "registerview": 0, "projectdownloadview": 1,
	"projectdeleteview": 1, "userdeleteview": 1, "userresult": 0,
	"projectresult": 0, "jpfconfigview": 1, "archiveview": 1,
	"projectview": 1, "addtest": 1, "addjpf": 1, "addproject": 1,
	"changeskeleton": 1, "submitarchive": 1, "login": 0, "register": 0,
	"logout": 1, "deleteproject": 1, "deleteuser": 1, "displaygraph": 0,
	"displayresult": 0, "getfiles": 0, "getsubmissions": 0,
	"skeleton.zip": 0, "index": 0, "favicon.ico": 0, "": 0, "statusview": 1,
	"createjpf": 1,
}

//GenerateViews is used to load all the basic views used by our web app.
func GenerateViews(router *pat.Router) {
	for name, view := range views {
		handleFunc := LoadView(name, view)
		lname := strings.ToLower(name)
		pattern := "/" + lname
		router.Add("GET", pattern, Handler(handleFunc)).Name(lname)
	}
}

//LoadView loads a view so that it is accessible in our web app.
func LoadView(name, view string) Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		ctx.Browse.View = views[name]
		args := map[string]interface{}{"ctx": ctx}
		return T(getNav(ctx), name).Execute(w, args)
	}
}

var posts = map[string]PostFunc{
	"addtest": AddTest, "addjpf": AddJPF,
	"addproject": AddProject, "changeskeleton": ChangeSkeleton,
	"submitarchive": SubmitArchive, "createjpf": CreateJPF,
}

//GeneratePosts loads post request handlers.
func GeneratePosts(router *pat.Router) {
	for name, fn := range posts {
		handleFunc := CreatePost(fn)
		pattern := "/" + name
		router.Add("POST", pattern, Handler(handleFunc)).Name(name)
	}
}

//CreatePost loads a post request handler.
func CreatePost(postFunc PostFunc) Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		err := postFunc(req, ctx)
		if err != nil {
			ctx.AddMessage("Could not complete submission.", true)
		} else {
			ctx.AddMessage("Successfully completed submission.", false)
		}
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
}

func deleteProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := DeleteProject(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not delete project.", true)
	} else {
		ctx.AddMessage("Successfully deleted project.", false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func deleteUser(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := DeleteUser(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not delete user.", true)
	} else {
		ctx.AddMessage("Successfully deleted user.", false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func downloadProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	path, err := LoadSkeleton(req)
	if err == nil {
		http.ServeFile(w, req, path)
	} else {
		ctx.AddMessage("Could not load project skeleton.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	}
	return err
}

func getSubmissions(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	subs, err := RetrieveSubmissions(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve submissions.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	var temp string
	if ctx.Browse.IsUser {
		temp = "userSubmissionResult"
	} else {
		temp = "projectSubmissionResult"
	}
	ctx.Browse.View = "home"
	return T(getNav(ctx), temp).Execute(w, map[string]interface{}{"ctx": ctx,
		"subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	fileinfo, err := RetrieveFileInfo(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve files.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	ctx.Browse.View = "home"
	return T(getNav(ctx), "fileResult").Execute(w,
		map[string]interface{}{"ctx": ctx, "fileinfo": fileinfo})
}

func displayResult(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	ctx.SetResult(req)
	ctx.Browse.View = "home"
	args, temps, err := analysisArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve results.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	return T(temps...).Execute(w, args)
}

func analysisArgs(req *http.Request, ctx *Context) (args map[string]interface{}, temps []string, err error) {
	files, err := RetrieveFiles(req, ctx)
	if err != nil {
		return
	}
	selected, err := getSelected(req, len(files)-1)
	if err != nil {
		return
	}
	curFile, err := getFile(files[selected].Id)
	if err != nil {
		return
	}
	projectId, err := util.ReadId(ctx.Browse.Pid)
	if err != nil {
		return
	}
	results, err := db.GetResultNames(projectId, true)
	if err != nil {
		return
	}
	curRes, err := GetResultData(ctx.Browse.Result, curFile.Id)
	if err != nil {
		return
	}
	neighbour, err := getNeighbour(req, len(files)-1)
	if err != nil{
		return
	}
	nextFile, err := getFile(files[neighbour].Id)
	if err != nil {
		return
	}
	nextRes, err := GetResultData(ctx.Browse.Result, nextFile.Id)
	if err != nil {
		return
	}
	args = map[string]interface{}{"ctx": ctx, "files": files,
		"selected": selected, "curFile":  curFile, 
		"curResult": curRes.GetData(), "results": results, 
		"nextFile": nextFile, "nextResult": nextRes.GetData(),
		"neighbour": neighbour} 
	temps = []string{getNav(ctx), "analysis", "pager", 
		curRes.Template(true), nextRes.Template(false)}
	return
}

func displayGraph(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	ctx.Browse.View = "home"
	ctx.SetResult(req)
	args, err := graphArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve graph data.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	temps := []string{getNav(ctx), "graphResult"}
	return T(temps...).Execute(w, args)
}

func graphArgs(req *http.Request, ctx *Context) (args map[string]interface{}, err error) {
	fileName, err := GetString(req, "filename")
	if err != nil {
		return
	}
	files, err := RetrieveFiles(req, ctx)
	if err != nil {
		return
	}
	projectId, err := util.ReadId(ctx.Browse.Pid)
	if err != nil {
		return
	}
	results, err := db.GetResultNames(projectId, false)
	results = append(results, "All")
	if err != nil {
		return
	}
	tipe, err := GetString(req, "type")
	if err != nil {
		return
	}
	graphArgs := LoadResultGraphData(ctx.Browse.Result, tipe, files)
	args = map[string]interface{}{
		"ctx": ctx, "files": files, "results": results, 
		"fileName": fileName, "graphArgs": graphArgs, "type": tipe,
	}
	return
}

func login(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := Login(req, ctx)
	if err != nil {
		ctx.AddMessage("Invalid username or password.", true)
	}
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return err
}

func register(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := Register(req, ctx)
	if err != nil {
		ctx.AddMessage("Invalid credentials.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	} else {
		ctx.AddMessage("Successfully registered.", false)
		http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	}
	return err
}

func logout(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return nil
}
