package webserver

import (
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"net/http"
	"strings"
)

var store sessions.Store

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
		util.Log(err)
	}
	ctx := NewContext(sess)
	buf := new(HttpBuffer)
	err = h(buf, req, ctx)
	if err != nil {
		util.Log(err)
	}
	if err = ctx.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.Apply(w)
}

var views = map[string]string{"homeView": "home", "testView": "submit",
	"skeletonView": "submit", "jpfFileView": "submit",
	"registerView": "register", "projectDownloadView": "download",
	"projectDeleteView": "delete", "userDeleteView": "delete",
	"userResult": "home", "projectResult": "home",
	"jpfConfigView": "submit", "archiveView": "submit",
	"projectView": "submit"}

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
		jsonData, err := loadGraphs(name)
		if err != nil{
			util.Log(err)
		}
		args := map[string]interface{}{"ctx": ctx, "jsonData": jsonData}
		return T(getNav(ctx), name).Execute(w, args)
	}
}

func loadGraphs(name string)(jsonData []map[string]interface{}, err error){
	switch name{
	case "projectResult":
		jsonData, err = loadProjectGraphData()
	case "userResult":
		jsonData, err = loadUserGraphData()
	}
	return
}

var posts = map[string]PostFunc{"addtest": AddTest, "addjpf": AddJPF,
	"addproject": AddProject, "changeskeleton": ChangeSkeleton,
	"submitarchive": SubmitArchive}

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
	names, err := RetrieveNames(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve files.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	ctx.Browse.View = "home"
	return T(getNav(ctx), "fileResult").Execute(w,
		map[string]interface{}{"ctx": ctx, "names": names})
}

func displayResult(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	args, temps, err := loadArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve results.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	return T(temps...).Execute(w, args)
}

func loadArgs(req *http.Request, ctx *Context) (args map[string]interface{}, temps []string, err error) {
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
	results, err := db.GetResultNames(projectId)
	if err != nil {
		return
	}
	ctx.SetResult(req)
	res, err := GetResultData(ctx.Browse.Result, curFile.Id)
	if err != nil {
		return
	}
	jsonData, err := loadResultGraphData(ctx.Browse.Result, curFile.Name, files)
	if err != nil {
		return
	}
	ctx.Browse.View = "home"
	args = map[string]interface{}{"ctx": ctx, "files": files,
		"selected": selected, "resultName": res.GetName(),
		"curFile": curFile, "curResult": res.GetData(),
		"results": results, "jsonData": jsonData}
	neighbour, ok := getNeighbour(req, len(files)-1)
	temps = []string{getNav(ctx), "fileInfo", res.Template(true)}
	if !ok {
		temps = append(temps, "singleResult")
		return
	}
	nextFile, err := getFile(files[neighbour].Id)
	if err != nil {
		return
	}
	res, err = GetResultData(ctx.Browse.Result, nextFile.Id)
	if err != nil {
		return
	}
	args["nextFile"] = nextFile
	args["nextResult"] = res.GetData()
	args["neighbour"] = neighbour
	temps = append(temps, "doubleResult", res.Template(false))
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
