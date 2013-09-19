//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package webserver

import (
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"net/http"
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
	err = CheckAccess(req.URL, ctx)
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

//getNav retrieves the navbar to display.
func getNav(ctx *Context) string {
	if _, err := ctx.Username(); err != nil {
		return "outNavbar"
	}
	return "inNavbar"
}

//downloadProject makes a project skeleton available for download.
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

//getSubmissions displays a list of submissions.
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

//configView loads a tool's configuration page.
func configView(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	tool, err := GetString(req, "tool")
	if err != nil {
		tool = "none"
	}
	ctx.Browse.View = "submit"
	return T(getNav(ctx), "configView", toolTemplate(tool)).Execute(w,
		map[string]interface{}{"ctx": ctx, "tool": tool})
}

//getFiles diplays information about files.
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

//displayResult displays a tool's result.
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

//analysisArgs loads arguments for displayResult
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

//displayGraph displays a graph for a tool's result.
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

//graphArgs loads arguments for displayGraph.
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

//login logs a user into the system.
func login(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	msg, err := Login(req, ctx)
	ctx.AddMessage(msg, err != nil)
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return err
}

//register registers a user with the system.
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

//Logout logs a user out of the system.
func Logout(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return nil
}
