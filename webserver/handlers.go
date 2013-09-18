package webserver

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"net/http"
	"net/url"
	"strings"
)

var (
	store sessions.Store
)

const (
	LOG_HANDLERS = "webserver/handlers.go"
)

type (
	//Handler is used to handle incoming requests.
	//It allows for better session management.
	Handler func(http.ResponseWriter, *http.Request, *Context) error
)

func init() {
	store = sessions.NewCookieStore(util.CookieKeys())
}

//ServeHTTP loads a the current session, handles  the request and
//then stores the session.
func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		util.Log(err, LOG_HANDLERS)
	}
	//Load our context from session
	ctx := LoadContext(sess)
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

//getNav
func getNav(ctx *Context) string {
	if _, err := ctx.Username(); err != nil {
		return "outNavbar"
	}
	return "inNavbar"
}

//checkAccess verifies that a user is allowed access to a url.
func checkAccess(url *url.URL, ctx *Context) error {
	//Rertieve the location they are requesting
	start := strings.LastIndex(url.Path, "/") + 1
	end := strings.Index(url.Path, "?")
	if end < 0 {
		end = len(url.Path)
	}
	if start > end {
		return fmt.Errorf("Invalid request %s", url.Path)
	}
	name := url.Path[start:end]
	perms := Permissions()
	//Get the permission and check it.
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

//downloadProject
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

//getSubmissions
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

//configView
func configView(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	tool, err := GetString(req, "tool")
	if err != nil {
		tool = "none"
	}
	ctx.Browse.View = "submit"
	return T(getNav(ctx), "configView", toolTemplate(tool)).Execute(w,
		map[string]interface{}{"ctx": ctx, "tool": tool})
}

//getFiles
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

//displayResult
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

//analysisArgs
func analysisArgs(req *http.Request, ctx *Context) (args map[string]interface{}, temps []string, err error) {
	files, err := RetrieveFiles(req, ctx)
	if err != nil {
		return
	}
	selected, serr := getSelected(req, len(files)-1)
	if serr == nil {
		ctx.Browse.Selected = selected
	}
	curFile, err := getFile(files[ctx.Browse.Selected].Id)
	if err != nil {
		return
	}
	results, err := db.ResultNames(ctx.Browse.Pid, true)
	if err != nil {
		return
	}
	curRes, err := GetResultData(ctx.Browse.ResultName, curFile.Id)
	if err != nil {
		return
	}
	neighbour, nerr := getNeighbour(req, len(files)-1)
	if nerr == nil {
		ctx.Browse.Next = neighbour
	}
	nextFile, err := getFile(files[ctx.Browse.Next].Id)
	if err != nil {
		return
	}
	nextRes, err := GetResultData(ctx.Browse.ResultName, nextFile.Id)
	if err != nil {
		return
	}
	currentLines := GetLines(req, "current")
	nextLines := GetLines(req, "next")
	args = map[string]interface{}{"ctx": ctx, "files": files,
		"curFile": curFile, "curResult": curRes.GetData(),
		"results": results, "nextFile": nextFile,
		"nextResult": nextRes.GetData(), "currentlines": currentLines,
		"nextlines": nextLines,
	}
	temps = []string{getNav(ctx), "analysis", "pager",
		tool.Template(curRes.GetName(), true), tool.Template(nextRes.GetName(), false)}
	return
}

//displayGraph
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

//graphArgs
func graphArgs(req *http.Request, ctx *Context) (args map[string]interface{}, err error) {
	fileName, ferr := GetString(req, "filename")
	if ferr == nil {
		ctx.Browse.FileName = fileName
	}
	files, err := RetrieveFiles(req, ctx)
	if err != nil {
		return
	}
	results, err := db.ResultNames(ctx.Browse.Pid, false)
	results = append(results, "All")
	if err != nil {
		return
	}
	tipe, err := GetString(req, "type")
	if err != nil {
		return
	}
	graphArgs := LoadResultGraphData(ctx.Browse.ResultName, tipe, files)
	args = map[string]interface{}{
		"ctx": ctx, "files": files, "results": results,
		"graphArgs": graphArgs, "type": tipe,
	}
	return
}

//login
func login(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	msg, err := Login(req, ctx)
	ctx.AddMessage(msg, err != nil)
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return err
}

//register
func register(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	msg, err := Register(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	} else {
		ctx.AddMessage(msg, false)
		http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	}
	return err
}

//logout
func logout(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return nil
}
