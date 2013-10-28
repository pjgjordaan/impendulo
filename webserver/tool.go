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
	"bytes"
	"errors"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

var (
	//Here we keep our tool configs' html template names.
	templates = map[string]string{
		jpf.NAME:        "jpfconfig",
		pmd.NAME:        "pmdconfig",
		junit.NAME:      "junitconfig",
		findbugs.NAME:   "findbugsconfig",
		checkstyle.NAME: "checkstyleconfig",
		"none":          "noconfig",
	}
)

//toolTemplate
func toolTemplate(tool string) string {
	return templates[tool]
}

//toolPermissions
func toolPermissions() map[string]user.Permission {
	return map[string]user.Permission{
		"createjpf":        user.TEACHER,
		"createpmd":        user.TEACHER,
		"createjunit":      user.TEACHER,
		"createfindbugs":   user.TEACHER,
		"createcheckstyle": user.TEACHER,
	}
}

//toolRequesters
func toolPosters() map[string]Poster {
	return map[string]Poster{
		"createpmd":        CreatePMD,
		"createjpf":        CreateJPF,
		"createjunit":      CreateJUnit,
		"createfindbugs":   CreateFindbugs,
		"createcheckstyle": CreateCheckstyle,
	}
}

//tools
func tools() []string {
	return []string{
		jpf.NAME, junit.NAME, pmd.NAME,
		findbugs.NAME, checkstyle.NAME, javac.NAME,
	}
}

//CreateCheckstyle
func CreateCheckstyle(req *http.Request, ctx *Context) (msg string, err error) {
	return
}

//CreateFindbugs
func CreateFindbugs(req *http.Request, ctx *Context) (msg string, err error) {
	return
}

//CreateJUnit adds a new JUnit test for a given project.
func CreateJUnit(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	testName, testBytes, err := ReadFormFile(req, "test")
	if err != nil {
		msg = "Could not read JUnit file."
		return
	}
	//A test does not always need data files.
	hasData := req.FormValue("data-check")
	var dataBytes []byte
	if hasData == "" {
		dataBytes = make([]byte, 0)
	} else if hasData == "true" {
		_, dataBytes, err = ReadFormFile(req, "data")
		if err != nil {
			msg = "Could not read data file."
			return
		}
	}
	//Read package name from file.
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	test := junit.NewTest(
		projectId, testName, pkg, testBytes, dataBytes,
	)
	err = db.AddJUnitTest(test)
	return
}

//AddJPF replaces a project's JPF configuration with a provided configuration file.
func AddJPF(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	_, data, err := ReadFormFile(req, "jpf")
	if err != nil {
		msg = "Could not read JPF configuration file."
		return
	}
	jpfConfig := jpf.NewConfig(projectId, data)
	err = db.AddJPFConfig(jpfConfig)
	return
}

//CreateJPF replaces a project's JPF configuration with a new, provided configuration.
func CreateJPF(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	//Read JPF properties.
	vals, err := readProperties(req)
	if err != nil {
		msg = err.Error()
		return
	}
	//Convert to JPF property file style.
	data, err := jpf.JPFBytes(vals)
	if err != nil {
		msg = "Could not create JPF configuration."
		return
	}
	//Save to db.
	jpfConfig := jpf.NewConfig(projectId, data)
	err = db.AddJPFConfig(jpfConfig)
	if err != nil {
		msg = "Could not create JPF configuration."
	} else {
		msg = "Successfully created JPF configuration."
	}
	return
}

//readProperties reads JPF properties from a raw string and stores them in a map.
func readProperties(req *http.Request) (vals map[string][]string, err error) {
	vals = make(map[string][]string)
	//Read configured listeners and search.
	listeners, getErr := GetStrings(req, "addedlisteners")
	if getErr == nil {
		vals["listener"] = listeners
	}
	search, getErr := GetString(req, "addedsearches")
	if getErr == nil {
		vals["search.class"] = []string{search}
	}
	other, getErr := GetString(req, "other")
	if getErr != nil {
		return
	}
	lines := strings.Split(other, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		err = readProperty(line, vals)
		if err != nil {
			return
		}
	}
	return
}

func readProperty(line string, props map[string][]string) (err error) {
	params := strings.Split(util.RemoveEmpty(line), "=")
	if len(params) != 2 {
		err = fmt.Errorf("Invalid JPF property %s.", line)
		return
	}
	key, val := params[0], params[1]
	if len(key) == 0 {
		err = errors.New("JPF key cannot be empty.")
		return
	}
	if len(val) == 0 {
		err = fmt.Errorf("JPF value for %s cannot be empty.", key)
		return
	}
	if !jpf.Allowed(key) {
		err = fmt.Errorf("Cannot set JPF property for %s.", key)
		return
	}
	split := strings.Split(val, ",")
	vals := make([]string, 0, len(split))
	for _, v := range split {
		if v != "" {
			vals = append(vals, v)
		}
	}
	if v, ok := props[key]; ok {
		props[key] = append(v, vals...)
	} else {
		props[key] = vals
	}
	return
}

//CreatePMD creates PMD rules for a project from a provided list.
func CreatePMD(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	rules, err := GetStrings(req, "ruleid")
	if err != nil {
		msg = "Could not read rules."
		return
	}
	pmdRules, err := pmd.NewRules(projectId, util.ToSet(rules))
	if err != nil {
		msg = "Could not create rules."
		return
	}
	err = db.AddPMDRules(pmdRules)
	if err != nil {
		msg = "Could not add rules."
	} else {
		msg = "Successfully added rules."
	}
	return
}

//RunTool runs a tool on submissions in a given project.
//Previous results are deleted if the user has specified that the tool
//should be rerun on all fi
func RunTool(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	all := req.FormValue("projectid") == "all"
	if err != nil && !all {
		return
	}
	tool, err := GetString(req, "tool")
	if err != nil {
		msg = "Could not read tool."
		return
	}
	var matcher bson.M
	if all {
		matcher = bson.M{}
	} else {
		matcher = bson.M{db.PROJECTID: projectId}
	}
	submissions, err := db.Submissions(matcher, bson.M{db.ID: 1})
	if err != nil {
		msg = "Could not retrieve submissions."
		return
	}
	var runAll bool
	if req.FormValue("runempty-check") == "true" {
		runAll = false
	} else {
		runAll = true
	}
	runTools(submissions, tool, runAll)
	msg = fmt.Sprintf("Successfully started running %s on project.", tool)
	return
}

func runTools(submissions []*project.Submission, tool string, runAll bool) {
	selector := bson.M{db.DATA: 0}
	for _, submission := range submissions {
		matcher := bson.M{db.SUBID: submission.Id}
		files, err := db.Files(matcher, selector)
		if err != nil {
			util.Log(err)
			continue
		}
		err = processing.StartSubmission(submission.Id)
		if err != nil {
			util.Log(err)
			continue
		}
		for _, file := range files {
			if runAll {
				if tool == "all" {
					removeAll(file)
				} else {
					removeOne(file, tool)
				}

			}
			err = processing.AddFile(file)
			if err != nil {
				util.Log(err)
			}
		}
		err = processing.EndSubmission(submission.Id)
		if err != nil {
			util.Log(err)
		}
	}
}

//removeAll removes all of the specified file's results from the db.
func removeAll(file *project.File) {
	for _, resId := range file.Results {
		id, ok := resId.(bson.ObjectId)
		if !ok {
			continue
		}
		err := db.RemoveById(db.RESULTS, id)
		if err != nil {
			util.Log(err)
		}
	}
	change := bson.M{
		db.SET: bson.M{
			db.RESULTS: bson.M{},
		},
	}
	err := db.Update(db.FILES, bson.M{db.ID: file.Id}, change)
	if err != nil {
		util.Log(err)
	}
}

//removeOne removes the specified file result from the db.
func removeOne(file *project.File, name string) {
	resultVal, ok := file.Results[name]
	if !ok {
		return
	}
	delete(file.Results, name)
	change := bson.M{
		db.SET: bson.M{
			db.RESULTS: file.Results,
		},
	}
	err := db.Update(db.FILES, bson.M{db.ID: file.Id}, change)
	if err != nil {
		util.Log(err)
	}
	resultId, ok := resultVal.(bson.ObjectId)
	if !ok {
		return
	}
	err = db.RemoveById(db.RESULTS, resultId)
	if err != nil {
		util.Log(err)
	}
	return
}

//GetResult retrieves a DisplayResult for a given file and result name.
func GetResult(resultName string, fileId bson.ObjectId) (res tool.DisplayResult, err error) {
	var file *project.File
	matcher := bson.M{db.ID: fileId}
	file, err = db.File(matcher, nil)
	if err != nil {
		return
	}
	switch resultName {
	case tool.CODE:
		res = tool.NewCodeResult(file.Data)
	case diff.NAME:
		res = diff.NewResult(file)
	case tool.SUMMARY:
		res = tool.NewSummaryResult()
		//Load summary for each available result.
		for name, resid := range file.Results {
			var currentRes tool.ToolResult
			currentRes, err = db.ToolResult(
				name, bson.M{
					db.ID: resid,
				}, nil,
			)
			if err != nil {
				return
			}
			res.(*tool.SummaryResult).AddSummary(currentRes)
		}
	default:
		ival, ok := file.Results[resultName]
		if !ok {
			res = tool.NewErrorResult(tool.NORESULT, resultName)
			return
		}
		switch val := ival.(type) {
		case bson.ObjectId:
			//Retrieve result from the db.
			matcher = bson.M{db.ID: val}
			res, err = db.DisplayResult(
				resultName, matcher, nil,
			)
		case string:
			//Error, so create new error result.
			res = tool.NewErrorResult(val, resultName)
		default:
			res = tool.NewErrorResult(tool.NORESULT, resultName)
		}
	}
	return
}
