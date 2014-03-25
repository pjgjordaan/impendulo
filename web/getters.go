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

package web

import (
	"code.google.com/p/gorilla/pat"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"net/http"
	"strings"
)

type (
	//Args represents arguments passed to html templates or to template.Execute.
	Args   map[string]interface{}
	Getter func(req *http.Request, ctx *Context) (Args, string, error)
)

var (
	getters map[string]Getter
)

//Getters retrieves all getters
func Getters() map[string]Getter {
	if getters == nil {
		getters = defaultGetters()
	}
	return getters
}

//defaultGetters loads the default getters.
func defaultGetters() map[string]Getter {
	return map[string]Getter{
		"configview": configView, "editdbview": editDBView,
		"loadproject": loadProject, "loadsubmission": loadSubmission,
		"loadfile": loadFile, "loaduser": loadUser,
		"displayresult": displayResult, "getfiles": getFiles, "displaychildresult": displayChildResult,
		"getsubmissionschart": getSubmissionsChart, "getsubmissions": getSubmissions,
	}
}

//GenerateGets loads post request handlers and adds them to the router.
func GenerateGets(router *pat.Router, gets map[string]Getter, views map[string]string) {
	for name, fn := range gets {
		handleFunc := fn.CreateGet(name, views[name])
		pattern := "/" + name
		router.Add("GET", pattern, Handler(handleFunc)).Name(name)
	}
}

func (this Getter) CreateGet(name, view string) Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		args, msg, e := this(req, ctx)
		if msg != "" {
			ctx.AddMessage(msg, e != nil)
		}
		if e != nil {
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return e
		}
		t, e := util.GetStrings(args, "templates")
		if e != nil {
			ctx.AddMessage("Could not load page.", true)
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return e
		}
		delete(args, "templates")
		ctx.Browse.View = view
		if ctx.Browse.View == "home" {
			ctx.Browse.SetLevel(name)
		}
		args["ctx"] = ctx
		return T(append(t, getNav(ctx))...).Execute(w, args)
	}
}

//configView loads a tool's configuration page.
func configView(req *http.Request, ctx *Context) (Args, string, error) {
	tool, e := GetString(req, "tool")
	if e != nil {
		tool = "none"
	}
	return Args{"tool": tool, "templates": []string{"configview", toolTemplate(tool)}},
		"", nil
}

//getSubmissions displays a list of submissions.
func getSubmissions(req *http.Request, ctx *Context) (Args, string, error) {
	e := ctx.Browse.Update(req)
	if e != nil {
		return nil, "Could not load submissions.", e
	}
	var matcher bson.M
	if !ctx.Browse.IsUser {
		matcher = bson.M{db.PROJECTID: ctx.Browse.Pid}
	} else {
		matcher = bson.M{db.USER: ctx.Browse.Uid}
	}
	subs, e := db.Submissions(matcher, nil, "-"+db.TIME)
	if e != nil {
		return nil, "Could not load submissions.", e
	}
	t := make([]string, 1)
	if ctx.Browse.IsUser {
		t[0] = "usersubmissionresult"
	} else {
		t[0] = "projectsubmissionresult"
	}
	return Args{"subRes": subs, "templates": t}, "", nil
}

//getFiles diplays information about files.
func getFiles(req *http.Request, ctx *Context) (Args, string, error) {
	e := ctx.Browse.Update(req)
	if e != nil {
		return nil, "Could not retrieve files.", e
	}
	matcher := bson.M{db.SUBID: ctx.Browse.Sid, db.OR: [2]bson.M{bson.M{db.TYPE: project.SRC}, bson.M{db.TYPE: project.TEST}}}
	fileInfo, e := db.FileInfos(matcher)
	if e != nil {
		return nil, "Could not retrieve files.", e
	}
	return Args{"fileInfo": fileInfo, "templates": []string{"fileresult"}}, "", nil
}

//editDBView
func editDBView(req *http.Request, ctx *Context) (Args, string, error) {
	editing, e := GetString(req, "editing")
	if e != nil {
		editing = "Project"
	}
	t := []string{"editdbview", "edit" + strings.ToLower(editing)}
	return Args{"editing": editing, "templates": t},
		"", nil
}

func loadProject(req *http.Request, ctx *Context) (Args, string, error) {
	projectId, msg, e := getProjectId(req)
	if e != nil {
		return nil, msg, e
	}
	p, e := db.Project(bson.M{db.ID: projectId}, nil)
	if e != nil {
		return nil, "Could not find project.", e
	}
	return Args{"editing": "Project", "project": p,
		"templates": []string{"editdbview", "editproject"}}, "", nil
}

func loadUser(req *http.Request, ctx *Context) (Args, string, error) {
	uname, msg, e := getUserId(req)
	if e != nil {
		return nil, msg, e
	}
	u, e := db.User(uname)
	if e != nil {
		return nil, fmt.Sprintf("Could not find user %s.", uname), e
	}
	return Args{"editing": "User", "user": u, "templates": []string{"editdbview", "edituser"}}, "", nil
}

func loadSubmission(req *http.Request, ctx *Context) (Args, string, error) {
	projectId, msg, e := getProjectId(req)
	if e != nil {
		return nil, msg, e
	}
	subId, msg, e := getSubId(req)
	if e != nil {
		return nil, msg, e
	}
	s, e := db.Submission(bson.M{db.ID: subId}, nil)
	if e != nil {
		return nil, "Could not find submission.", e
	}
	return Args{"editing": "Submission", "projectId": projectId, "submission": s,
		"templates": []string{"editdbview", "editsubmission"}}, "", nil
}

func loadFile(req *http.Request, ctx *Context) (Args, string, error) {
	projectId, msg, e := getProjectId(req)
	if e != nil {
		return nil, msg, e
	}
	subId, msg, e := getSubId(req)
	if e != nil {
		return nil, msg, e
	}
	fileId, msg, e := getFileId(req)
	if e != nil {
		return nil, msg, e
	}
	f, e := db.File(bson.M{db.ID: fileId}, nil)
	if e != nil {
		return nil, "Could not find file.", e
	}
	return Args{"editing": "File", "projectId": projectId, "submissionId": subId, "file": f,
		"templates": []string{"editdbview", "editfile"}}, "", nil
}

//displayResult displays a tool's result.
func displayResult(req *http.Request, ctx *Context) (Args, string, error) {
	a, e := _displayResult(req, ctx)
	if e != nil {
		return nil, "Could not load results.", e
	}
	return a, "", nil
}

func _displayResult(req *http.Request, ctx *Context) (Args, error) {
	e := ctx.Browse.Update(req)
	if e != nil {
		return nil, e
	}
	if ctx.Browse.childResult() {
		return _displayChildResult(req, ctx)
	}
	files, e := Snapshots(ctx.Browse.Sid, ctx.Browse.File, ctx.Browse.Type)
	if e != nil {
		return nil, e
	}
	currentFile, e := getFile(files[ctx.Browse.Current].Id)
	if e != nil {
		return nil, e
	}
	results, e := analysisNames(ctx.Browse.Pid, ctx.Browse.Type)
	if e != nil {
		return nil, e
	}
	currentResult, e := GetResult(ctx.Browse.Result, currentFile.Id)
	if e != nil {
		return nil, e
	}
	nextFile, e := getFile(files[ctx.Browse.Next].Id)
	if e != nil {
		return nil, e
	}
	nextResult, e := GetResult(ctx.Browse.Result, nextFile.Id)
	if e != nil {
		return nil, e
	}
	t := []string{"analysisview", "pager", "srcanalysis", ""}
	if !isError(currentResult) || isError(nextResult) {
		t[3] = currentResult.Template()
	} else {
		t[3] = nextResult.Template()
	}
	return Args{
		"files": files, "currentFile": currentFile, "currentResult": currentResult, "results": results,
		"nextFile": nextFile, "nextResult": nextResult, "templates": t,
	}, nil
}

func displayChildResult(req *http.Request, ctx *Context) (Args, string, error) {
	a, e := _displayChildResult(req, ctx)
	if e != nil {
		return nil, "Could not load results.", e
	}
	return a, "", nil
}

func _displayChildResult(req *http.Request, ctx *Context) (Args, error) {
	e := ctx.Browse.Update(req)
	if e != nil {
		return nil, e
	}
	parentFiles, e := Snapshots(ctx.Browse.Sid, ctx.Browse.File, ctx.Browse.Type)
	if e != nil {
		return nil, e
	}
	if cur, ce := getCurrent(req, len(parentFiles)-1); ce == nil {
		ctx.Browse.Current = cur
	}
	if next, ne := getNext(req, len(parentFiles)-1); ne == nil {
		ctx.Browse.Next = next
	}
	currentFile, e := getFile(parentFiles[ctx.Browse.Current].Id)
	if e != nil {
		return nil, e
	}
	results, e := resultNames(ctx.Browse.Pid, testView)
	if e != nil {
		return nil, e
	}
	nextFile, e := getFile(parentFiles[ctx.Browse.Next].Id)
	if e != nil {
		return nil, e
	}
	childFiles, e := Snapshots(ctx.Browse.Sid, ctx.Browse.ChildFile, ctx.Browse.ChildType)
	if e != nil {
		return nil, e
	}
	currentChild, e := db.File(bson.M{db.ID: childFiles[ctx.Browse.CurrentChild].Id}, nil)
	if e != nil {
		return nil, e
	}
	currentChildResult, e := GetTestResult(ctx.Browse.Result, currentChild.Id.Hex(), currentFile.Id)
	if e != nil {
		return nil, e
	}
	nextChildResult, e := GetTestResult(ctx.Browse.Result, currentChild.Id.Hex(), nextFile.Id)
	if e != nil {
		return nil, e
	}
	t := []string{"analysisview", "pager", "testanalysis", ""}
	if !isError(currentChildResult) || isError(nextChildResult) {
		t[3] = currentChildResult.Template()
	} else {
		t[3] = nextChildResult.Template()
	}
	return Args{
		"files": parentFiles, "childFiles": childFiles, "currentFile": currentFile, "nextFile": nextFile,
		"results": results, "childFile": currentChild, "currentChildResult": currentChildResult,
		"nextChildResult": nextChildResult, "templates": t,
	}, nil
}

//getSubmissionsChart displays a chart of submissions.
func getSubmissionsChart(req *http.Request, ctx *Context) (Args, string, error) {
	e := ctx.Browse.Update(req)
	if e != nil {
		return nil, "Could not load chart.", e
	}
	var matcher bson.M
	if !ctx.Browse.IsUser {
		matcher = bson.M{db.PROJECTID: ctx.Browse.Pid}
	} else {
		matcher = bson.M{db.USER: ctx.Browse.Uid}
	}
	subs, e := db.Submissions(matcher, nil, "-"+db.TIME)
	if e != nil {
		return nil, "Could not load chart.", e
	}
	chartData := SubmissionChart(subs)
	t := make([]string, 1)
	if ctx.Browse.IsUser {
		t[0] = "usersubmissionchart"
	} else {
		t[0] = "projectsubmissionchart"
	}
	return Args{"chart": chartData, "templates": t}, "", nil
}
