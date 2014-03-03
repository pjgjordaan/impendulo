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
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
)

type (
	Temps  []string
	Getter func(req *http.Request, ctx *Context) (args Args, templates Temps, msg string, err error)
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
		"loadfile": loadFile, "loaduser": loadUser, "displaychart": displayChart,
		"displayresult": displayResult, "getfiles": getFiles, "displaytestresult": displayTestResult,
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
		args, templates, msg, err := this(req, ctx)
		if msg != "" {
			ctx.AddMessage(msg, err != nil)
		}
		if err != nil {
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return err
		}
		ctx.Browse.View = view
		if ctx.Browse.View == "home" {
			ctx.Browse.SetLevel(name)
		}
		args["ctx"] = ctx
		return T(append(templates, getNav(ctx))...).Execute(w, args)
	}
}

//configView loads a tool's configuration page.
func configView(req *http.Request, ctx *Context) (Args, Temps, string, error) {
	tool, err := GetString(req, "tool")
	if err != nil {
		tool = "none"
	}
	return Args{"tool": tool},
		Temps{"configview", toolTemplate(tool)},
		"", nil
}

//getSubmissions displays a list of submissions.
func getSubmissions(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	subs, err := RetrieveSubmissions(req, ctx)
	if err != nil {
		msg = "Could not retrieve submissions."
		return
	}
	t = make(Temps, 1)
	if ctx.Browse.IsUser {
		t[0] = "usersubmissionresult"
	} else {
		t[0] = "projectsubmissionresult"
	}
	a = Args{"subRes": subs}
	return
}

//getFiles diplays information about files.
func getFiles(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	sid, msg, err := getSubId(req)
	if err != nil {
		return
	}
	matcher := bson.M{db.SUBID: sid, db.OR: [2]bson.M{bson.M{db.TYPE: project.SRC}, bson.M{db.TYPE: project.TEST}}}
	fileInfo, err := db.FileInfos(matcher)
	if err != nil {
		msg = "Could not retrieve files."
		return
	}
	sub, err := db.Submission(
		bson.M{db.ID: sid},
		bson.M{db.PROJECTID: 1, db.USER: 1},
	)
	if err != nil {
		msg = "Could not retrieve submission."
		return
	}
	ctx.Browse.Pid = sub.ProjectId
	ctx.Browse.Uid = sub.User
	ctx.Browse.Sid = sid
	a = Args{"fileInfo": fileInfo}
	t = Temps{"fileresult"}
	return
}

//editDBView
func editDBView(req *http.Request, ctx *Context) (Args, Temps, string, error) {
	editing, err := GetString(req, "editing")
	if err != nil {
		editing = "Project"
	}
	return Args{"editing": editing},
		Temps{"editdbview", "edit" + strings.ToLower(editing)},
		"", nil
}

func loadProject(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	p, err := db.Project(bson.M{db.ID: projectId}, nil)
	if err != nil {
		msg = "Could not find project."
		return
	}
	a = Args{"editing": "Project", "project": p}
	t = Temps{"editdbview", "editproject"}
	return
}

func loadUser(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	uname, msg, err := getUserId(req)
	if err != nil {
		return
	}
	u, err := db.User(uname)
	if err != nil {
		msg = fmt.Sprintf("Could not find user %s.", uname)
		return
	}
	a = Args{"editing": "User", "user": u}
	t = Temps{"editdbview", "edituser"}
	return
}

func loadSubmission(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	a = Args{"editing": "Submission"}
	t = Temps{"editdbview", "editsubmission"}
	a["projectId"], msg, err = getProjectId(req)
	if err != nil {
		return
	}
	subId, _, serr := getSubId(req)
	if serr != nil {
		return
	}
	a["submission"], err = db.Submission(bson.M{db.ID: subId}, nil)
	if err != nil {
		msg = "Could not find submission."
		return
	}
	return
}

func loadFile(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	a = Args{"editing": "File"}
	t = Temps{"editdbview", "editfile"}
	a["projectId"], msg, err = getProjectId(req)
	if err != nil {
		return
	}
	subId, _, serr := getSubId(req)
	if serr != nil {
		return
	}
	a["submissionId"] = subId
	fileId, _, ferr := getFileId(req)
	if ferr != nil {
		return
	}
	a["file"], err = db.File(bson.M{db.ID: fileId}, nil)
	if err != nil {
		msg = "Could not find file."
		return
	}
	return
}

func displayChart(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	defer func() {
		if err != nil {
			msg = "Could not load chart."
		}
	}()
	err = SetContext(req, ctx.Browse.SetUid, ctx.Browse.SetSid,
		ctx.Browse.SetResult, ctx.Browse.SetFile)
	if err != nil {
		return
	}
	files, err := Snapshots(ctx.Browse.Sid, ctx.Browse.File, ctx.Browse.Type)
	if err != nil {
		return
	}
	results, err := db.ChartResultNames(ctx.Browse.Pid)
	if err != nil {
		return
	}
	chart, err := LoadChart(ctx.Browse.Result, files)
	if err != nil {
		return
	}
	compStr := ""
	compId, cerr := util.ReadId(req.FormValue("compare"))
	if cerr == nil {
		var compFiles []*project.File
		compFiles, err = Snapshots(compId, ctx.Browse.File, ctx.Browse.Type)
		if err != nil {
			return
		}
		var compChart ChartData
		compChart, err = LoadChart(ctx.Browse.Result, compFiles)
		if err != nil {
			return
		}
		chart = append(chart, compChart...)
		compStr = compId.Hex()
	}
	a = Args{
		"files": files, "results": results,
		"chart": chart, "compare": compStr,
	}
	t = Temps{"charts"}
	return
}

//displayResult displays a tool's result.
func displayResult(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	defer func() {
		if err != nil {
			msg = "Could not load results."
		}
	}()
	err = SetContext(req, ctx.Browse.SetUid, ctx.Browse.SetSid, ctx.Browse.SetResult, ctx.Browse.SetFile, ctx.Browse.SetDisplayCount)
	if err != nil {
		return
	}
	if ctx.Browse.Type == project.TEST {
		return displayTestResult(req, ctx)
	}
	files, err := Snapshots(ctx.Browse.Sid, ctx.Browse.File, ctx.Browse.Type)
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
				ctx.Browse.Next = (index + 1) % len(files)
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
	results, err := db.AllResultNames(ctx.Browse.Pid)
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
		var bug *tool.Bug
		bug, err = codeBug(req)
		if err != nil {
			return
		} else if bug != nil {
			switch bug.FileId {
			case currentFile.Id:
				currentResult.(*tool.CodeResult).Bug = bug
			case nextFile.Id:
				nextResult.(*tool.CodeResult).Bug = bug
			default:
				err = fmt.Errorf("Result %s does not match any files.", bug.FileId.Hex())
				return
			}
		}
	}
	a = Args{
		"files": files, "currentFile": currentFile,
		"currentResult": currentResult, "results": results,
		"nextFile": nextFile, "nextResult": nextResult,
	}
	t = Temps{"analysis", "pager", "", "noadditional"}
	if !isError(currentResult) || isError(nextResult) {
		t[2] = currentResult.Template()
	} else {
		t[2] = nextResult.Template()
	}
	return
}

func displayTestResult(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	defer func() {
		if err != nil {
			msg = "Could not load results."
		}
	}()
	parentFiles, err := Snapshots(ctx.Browse.Sid, ctx.Browse.File, ctx.Browse.Type)
	if err != nil {
		return
	}
	if cur, cerr := getCurrent(req, len(parentFiles)-1); cerr == nil {
		ctx.Browse.Current = cur
	}
	if next, nerr := getNext(req, len(parentFiles)-1); nerr == nil {
		ctx.Browse.Next = next
	}
	currentFile, err := getFile(parentFiles[ctx.Browse.Current].Id)
	if err != nil {
		return
	}
	results, err := resultNames(ctx.Browse.Pid, project.TEST, true)
	if err != nil {
		return
	}
	currentResult, err := GetResult(ctx.Browse.Result, currentFile.Id)
	if err != nil {
		return
	}
	nextFile, err := getFile(parentFiles[ctx.Browse.Next].Id)
	if err != nil {
		return
	}
	nextResult, err := GetResult(ctx.Browse.Result, nextFile.Id)
	if err != nil {
		return
	}
	if ctx.Additional.File == "" {
		ctx.Additional.File, err = testedFileName(ctx.Browse.Sid)
		if err != nil {
			return
		}
	}
	childFiles, err := Snapshots(ctx.Browse.Sid, ctx.Additional.File, ctx.Additional.Type)
	if err != nil {
		return
	}
	if cur, cerr := getIndex(req, "currentchild", len(childFiles)-1); cerr == nil {
		ctx.Additional.Current = cur
	}
	currentChild, err := db.File(bson.M{db.ID: childFiles[ctx.Additional.Current].Id}, nil)
	if err != nil {
		return
	}
	currentChildResult, err := GetResult(currentChild.Id.Hex(), currentFile.Id)
	if err != nil {
		return
	}
	nextChildResult, err := GetResult(currentChild.Id.Hex(), nextFile.Id)
	if err != nil {
		return
	}
	a = Args{
		"files": parentFiles, "childFiles": childFiles, "currentFile": currentFile,
		"currentResult": currentResult, "nextFile": nextFile, "nextResult": nextResult,
		"results": results, "childFile": currentChild, "currentChildResult": currentChildResult,
		"nextChildResult": nextChildResult,
	}
	t = Temps{"analysis", "pager", "", "additionalanalysis", "junitadditional"}
	if !isError(currentResult) || isError(nextResult) {
		t[2] = currentResult.Template()
	} else {
		t[2] = nextResult.Template()
	}
	return
}

//getSubmissionsChart displays a chart of submissions.
func getSubmissionsChart(req *http.Request, ctx *Context) (a Args, t Temps, msg string, err error) {
	subs, err := RetrieveSubmissions(req, ctx)
	if err != nil {
		msg = "Could not retrieve submissions."
		return
	}
	chartData := SubmissionChart(subs)
	a = Args{"chart": chartData}
	t = make(Temps, 1)
	if ctx.Browse.IsUser {
		t[0] = "usersubmissionchart"
	} else {
		t[0] = "projectsubmissionchart"
	}
	return
}
