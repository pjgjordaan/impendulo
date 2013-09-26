//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
	neighbour, nerr := getNeighbour(req, len(files)-1)
	if nerr == nil {
		ctx.Browse.Next = neighbour
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

//displayChart displays a chart for a tool's result.
func displayChart(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	ctx.Browse.View = "home"
	ctx.SetResult(req)
	args, err := chartArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve chart data.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	temps := []string{getNav(ctx), "charts"}
	return T(temps...).Execute(w, args)
}

//chartArgs loads arguments for displaychart.
func chartArgs(req *http.Request, ctx *Context) (args map[string]interface{}, err error) {
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
	chart := LoadChart(ctx.Browse.ResultName, tipe, files)
	args = map[string]interface{}{
		"ctx": ctx, "files": files, "results": results,
		"chart": chart, "type": tipe,
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
