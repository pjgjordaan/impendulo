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
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
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
	uname, err := ctx.Username()
	if err != nil {
		return "outNavbar"
	}
	u, err := db.User(uname)
	if err != nil {
		return "outNavbar"
	}
	switch u.Access {
	case user.STUDENT:
		return "studentNavbar"
	case user.TEACHER:
		return "teacherNavbar"
	case user.ADMIN:
		return "adminNavbar"
	}
	return "outNavbar"
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

//configView loads a tool's configuration page.
func configView(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	tool, err := GetString(req, "tool")
	if err != nil {
		tool = "none"
	}
	ctx.Browse.View = "submit"
	args := map[string]interface{}{"ctx": ctx, "tool": tool}
	return T(getNav(ctx), "configView", toolTemplate(tool)).Execute(w, args)
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
	ctx.Browse.Level = SUBMISSIONS
	args := map[string]interface{}{"ctx": ctx, "subRes": subs}
	return T(getNav(ctx), temp).Execute(w, args)
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
	ctx.Browse.Level = FILES
	args := map[string]interface{}{"ctx": ctx, "fileInfo": fileinfo}
	return T(getNav(ctx), "fileResult").Execute(w, args)
}

//showResult displays a tool's result.
func showResult(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	args, temps, err := analysisArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve results.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	ctx.Browse.View = "home"
	ctx.Browse.Level = ANALYSIS
	return T(temps...).Execute(w, args)
}

//analysisArgs loads arguments for displayResult
func analysisArgs(req *http.Request, ctx *Context) (args map[string]interface{}, temps []string, err error) {
	oldName := ctx.Browse.Result
	err = SetContext(req, ctx.Browse.SetUid, ctx.Browse.SetSid, ctx.Browse.SetResult, ctx.Browse.SetFile)
	if err != nil {
		return
	}
	files, err := Snapshots(ctx.Browse.Sid, ctx.Browse.File)
	if err != nil {
		return
	}
	ctx.Browse.Current, err = getCurrent(req, len(files)-1)
	if err != nil {
		time, terr := strconv.ParseInt(req.FormValue("time"), 10, 64)
		if terr != nil {
			return
		}
		found := false
		for index, file := range files {
			if file.Time == time {
				ctx.Browse.Current = index
				ctx.Browse.Next = index + 1
				found = true
				break
			}
		}
		if !found {
			return
		}
	} else {
		ctx.Browse.Next, err = getNext(req, len(files)-1)
		if err != nil {
			return
		}
	}
	currentFile, err := getFile(files[ctx.Browse.Current].Id)
	if err != nil {
		return
	}
	results, err := db.ResultNames(ctx.Browse.Pid, true)
	if err != nil {
		return
	}
	currentResult, err := GetResult(ctx.Browse.Result, currentFile.Id)
	if err != nil {
		return
	}
	nextFile, err := getFile(files[ctx.Browse.Next].Id)
	if err != nil {
		return
	}
	nextResult, err := GetResult(ctx.Browse.Result, nextFile.Id)
	if err != nil {
		return
	}
	if ctx.Browse.Result == tool.CODE {
		lines, id := highlightArgs(req, oldName)
		if lines != nil {
			if id == currentFile.Id {
				currentResult.(*tool.CodeResult).Lines = lines
			} else if id == nextFile.Id {
				nextResult.(*tool.CodeResult).Lines = lines
			} else {
				err = fmt.Errorf("Caller %v does not match any files.", id)
				return
			}
		}
	}
	args = map[string]interface{}{
		"ctx": ctx, "files": files, "currentFile": currentFile,
		"currentResult": currentResult, "results": results,
		"nextFile": nextFile, "nextResult": nextResult,
	}
	var template string
	if !isError(currentResult) || isError(nextResult) {
		template = currentResult.Template()
	} else {
		template = nextResult.Template()
	}
	temps = []string{
		getNav(ctx), "analysis", "pager", template,
	}
	return
}

//highlightArgs loads the file's id which we want to highlight the lines in as well as the lines.
func highlightArgs(req *http.Request, name string) (lines []int, id bson.ObjectId) {
	resId, err := util.ReadId(req.FormValue("caller"))
	if err != nil {
		return
	}
	result, err := db.ToolResult(name, bson.M{project.ID: resId}, bson.M{project.FILEID: 1})
	if err != nil {
		util.Log(err)
		return
	}
	id = result.GetFileId()
	lines, err = GetLines(req)
	if err != nil {
		util.Log(err)
	}
	return
}

//showChart displays a chart for a tool's result.
func showChart(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	args, err := chartArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve chart data.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	ctx.Browse.View = "home"
	ctx.Browse.Level = CHART
	temps := []string{getNav(ctx), "charts"}
	return T(temps...).Execute(w, args)
}

//chartArgs loads arguments for displaychart.
func chartArgs(req *http.Request, ctx *Context) (args map[string]interface{}, err error) {
	err = SetContext(req, ctx.Browse.SetUid, ctx.Browse.SetSid, ctx.Browse.SetResult, ctx.Browse.SetFile)
	if err != nil {
		return
	}
	files, err := Snapshots(ctx.Browse.Sid, ctx.Browse.File)
	if err != nil {
		return
	}
	results, err := db.ResultNames(ctx.Browse.Pid, false)
	if err != nil {
		return
	}
	startTime := files[0].Time
	chart := LoadChart(ctx.Browse.Result, files, startTime).Data
	compareStr := ""
	compareId, cerr := util.ReadId(req.FormValue("compare"))
	if cerr == nil {
		var compareFiles []*project.File
		compareFiles, err = Snapshots(compareId, ctx.Browse.File)
		if err != nil {
			return
		}
		compareChart := LoadChart(ctx.Browse.Result, compareFiles, startTime).Data
		chart = append(chart, compareChart...)
		compareStr = compareId.Hex()
	}
	args = map[string]interface{}{
		"ctx": ctx, "files": files, "results": results,
		"chart": chart, "compare": compareStr,
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
