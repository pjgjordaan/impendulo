package webserver

import (
	"code.google.com/p/gorilla/pat"
	"net/http"
	"strings"
)

var (
	postFuncs   map[string]PostFunc
	viewRoutes  map[string]string
	permissions map[string]int
)

//A function used to add data to the database.
type PostFunc func(*http.Request, *Context) error

func PostFuncs() map[string]PostFunc {
	if postFuncs != nil {
		return postFuncs
	}
	postFuncs = toolPostFuncs()
	defualt := defaultPostFuncs()
	for k, v := range defualt {
		postFuncs[k] = v
	}
	return postFuncs
}

func defaultPostFuncs() map[string]PostFunc {
	return map[string]PostFunc{
		"addproject": AddProject, "changeskeleton": ChangeSkeleton,
		"submitarchive": SubmitArchive, "runtool": RunTool,
	}
}

//GeneratePosts loads post request handlers.
func GeneratePosts(router *pat.Router) {
	posts := PostFuncs()
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

func Views() map[string]string {
	if viewRoutes != nil {
		return viewRoutes
	}
	viewRoutes = defaultViews()
	return viewRoutes
}

func defaultViews() map[string]string {
	return map[string]string{
		"homeView": "home", "skeletonView": "submit",
		"registerView": "register", "projectDownloadView": "download",
		"projectDeleteView": "delete", "userDeleteView": "delete",
		"userResult": "home", "projectResult": "home",
		"archiveView": "submit", "projectView": "submit",
		"statusView": "status", "runToolView": "tool",
	}
}

func Permissions() map[string]int {
	if permissions != nil {
		return permissions
	}
	permissions = toolPermissions()
	defualt := defaultPermissions()
	for k, v := range defualt {
		permissions[k] = v
	}
	return permissions
}

func defaultPermissions() map[string]int {
	return map[string]int{
		"homeview": 1, "skeletonview": 1, "configview": 1,
		"registerview": 0, "projectdownloadview": 1,
		"projectdeleteview": 1, "userdeleteview": 1, "userresult": 0,
		"projectresult": 0, "archiveview": 1, "projectview": 1,
		"addproject": 1, "changeskeleton": 1, "submitarchive": 1,
		"login": 0, "register": 0, "logout": 1, "deleteproject": 1,
		"deleteuser": 1, "displaygraph": 0, "displayresult": 0, "getfiles": 0,
		"getsubmissions": 0, "skeleton.zip": 0, "index": 0, "favicon.ico": 0,
		"": 0, "statusview": 1, "runtoolview": 1, "runtool": 1,
	}
}

//GenerateViews is used to load all the basic views used by our web app.
func GenerateViews(router *pat.Router) {
	views := Views()
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
		views := Views()
		ctx.Browse.View = views[name]
		args := map[string]interface{}{"ctx": ctx}
		return T(getNav(ctx), name).Execute(w, args)
	}
}
