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
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	mk "github.com/godfried/impendulo/tool/make"
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
		mk.NAME:         "makeconfig",
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
		"createmake":       user.TEACHER,
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
		"createmake":       CreateMake,
	}
}

func configTools() []string {
	return []string{junit.NAME, jpf.NAME, pmd.NAME, findbugs.NAME, checkstyle.NAME, mk.NAME}
}

//tools
func tools(projectId bson.ObjectId) (ret []string, err error) {
	project, err := db.Project(bson.M{db.ID: projectId}, nil)
	if err != nil {
		return
	}
	switch tool.Language(project.Lang) {
	case tool.JAVA:
		ret = []string{pmd.NAME, findbugs.NAME, checkstyle.NAME, javac.NAME}
		_, jerr := db.JPFConfig(bson.M{db.PROJECTID: projectId}, bson.M{db.ID: 1})
		if jerr == nil {
			ret = append(ret, jpf.NAME)
		}
		tests, terr := db.JUnitTests(bson.M{db.PROJECTID: projectId}, bson.M{db.NAME: 1})
		if terr == nil {
			for _, test := range tests {
				name, _ := util.Extension(test.Name)
				ret = append(ret, name)
			}
		}
		if db.Contains(db.TESTS, bson.M{db.PROJECTID: projectId, db.TYPE: junit.USER}) {
			ret = append(ret, jacoco.NAME)
		}
	case tool.C:
		ret = []string{mk.NAME, gcc.NAME}
	default:
		err = fmt.Errorf("Unknown language %s.", project.Lang)
	}
	return
}

//CreateCheckstyle
func CreateCheckstyle(req *http.Request, ctx *Context) (msg string, err error) {
	return
}

//CreateFindbugs
func CreateFindbugs(req *http.Request, ctx *Context) (msg string, err error) {
	return
}

func CreateMake(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	_, data, err := ReadFormFile(req, "makefile")
	if err != nil {
		msg = "Could not read Makefile."
		return
	}
	makefile := mk.NewMakefile(projectId, data)
	err = db.AddMakefile(makefile)
	if err != nil {
		msg = "Could not create Makefile."
	} else {
		msg = "Successfully created Makefile."
	}
	return
}

//CreateJUnit adds a new JUnit test for a given project.
func CreateJUnit(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	tipe, msg, err := getTestType(req)
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
	test := junit.NewTest(projectId, testName, pkg, tipe, testBytes, dataBytes)
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

//RunTools runs a tool on submissions in a given project.
//Previous results are deleted if the user has specified that the tool
//should be rerun on all fi
func RunTools(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, msg, err := getProjectId(req)
	if err != nil {
		return
	}
	tools, err := GetStrings(req, "tools")
	if err != nil {
		msg = "Could not read tool."
		return
	}
	users, err := GetStrings(req, "users")
	if err != nil {
		msg = "Could not read tool."
		return
	}
	matcher := bson.M{db.PROJECTID: projectId, db.USER: bson.M{db.IN: users}}
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
	childTools := make([]string, 0, len(tools))
	baseTools := make([]string, 0, len(tools))
	for _, t := range tools {
		if isChildTool(t, projectId) {
			childTools = append(childTools, t)
		} else {
			baseTools = append(baseTools, t)
		}
	}
	redoSubmissions(submissions, baseTools, childTools, runAll)
	msg = "Successfully started running tools on submissions."
	return
}

func isChildTool(name string, projectId bson.ObjectId) bool {
	return name == jacoco.NAME || db.Contains(db.TESTS, bson.M{db.PROJECTID: projectId, db.NAME: name})
}

func redoSubmissions(submissions []*project.Submission, tools, childTools []string, runAll bool) {
	selector := bson.M{db.DATA: 0}
	for _, submission := range submissions {
		var srcs []*project.File
		var err error
		if len(childTools) > 0 {
			srcs, err = db.Files(bson.M{db.SUBID: submission.Id, db.TYPE: project.SRC}, bson.M{db.ID: 1})
			if err != nil {
				util.Log(err)
			}
		}
		matcher := bson.M{db.SUBID: submission.Id}
		files, err := db.Files(matcher, selector)
		if err != nil {
			util.Log(err)
			continue
		}
		for _, file := range files {
			runTools(file, tools, runAll)
			if file.Type == project.TEST && len(childTools) > 0 {
				runChildTools(file, srcs, childTools, runAll)
			}
		}
		err = processing.RedoSubmission(submission.Id)
		if err != nil {
			util.Log(err)
		}
	}
}

func runTools(file *project.File, tools []string, runAll bool) (err error) {
	for _, t := range tools {
		resultVal, ok := file.Results[t]
		if !ok {
			continue
		}
		resultId, isId := resultVal.(bson.ObjectId)
		if !runAll && isId {
			continue
		}
		delete(file.Results, t)
		if !isId {
			continue
		}
		err = db.RemoveById(db.RESULTS, resultId)
		if err != nil {
			util.Log(err)
		}
	}
	change := bson.M{db.SET: bson.M{db.RESULTS: file.Results}}
	err = db.Update(db.FILES, bson.M{db.ID: file.Id}, change)
	if err != nil {
		util.Log(err)
	}
	return
}

func runChildTools(test *project.File, srcs []*project.File, childTools []string, runAll bool) (err error) {
	for _, s := range srcs {
		for _, t := range childTools {
			key := t + "-" + s.Id.Hex()
			resultVal, ok := test.Results[key]
			if !ok {
				continue
			}
			resultId, isId := resultVal.(bson.ObjectId)
			if !runAll && isId {
				continue
			}
			delete(test.Results, key)
			if !isId {
				continue
			}
			err = db.RemoveById(db.RESULTS, resultId)
			if err != nil {
				util.Log(err)
			}
		}
	}
	change := bson.M{db.SET: bson.M{db.RESULTS: test.Results}}
	err = db.Update(db.FILES, bson.M{db.ID: test.Id}, change)
	if err != nil {
		util.Log(err)
	}
	return
}

//GetResult retrieves a DisplayResult for a given file and result name.
func GetResult(name string, fileId bson.ObjectId) (tool.DisplayResult, error) {
	m := bson.M{db.ID: fileId}
	f, err := db.File(m, nil)
	if err != nil {
		return nil, err
	}
	s, err := db.Submission(bson.M{db.ID: f.SubId}, nil)
	if err != nil {
		return nil, err
	}
	switch name {
	case tool.CODE:
		p, err := db.Project(bson.M{db.ID: s.ProjectId}, nil)
		if err != nil {
			return nil, err
		}
		return tool.NewCodeResult(p.Lang, f.Data), nil
	case diff.NAME:
		return diff.NewResult(f), nil
	case tool.SUMMARY:
		r := tool.NewSummaryResult()
		//Load summary for each available result.
		for _, id := range f.Results {
			cur, err := db.ToolResult(bson.M{db.ID: id}, nil)
			if err != nil {
				return nil, err
			}
			r.AddSummary(cur)
		}
		return r, nil
	default:
		ival, ok := f.Results[name]
		if !ok {
			if strings.HasPrefix(junit.NAME, name) {
				name = "unit tests"
			} else if strings.HasPrefix(jacoco.NAME, name) {
				name = "jacoco test coverage"
			}
			return tool.NewErrorResult(tool.NORESULT, name), nil
		}
		switch v := ival.(type) {
		case bson.ObjectId:
			//Retrieve result from the db.
			return db.DisplayResult(bson.M{db.ID: v}, nil)
		case string:
			//Error, so create new error result.
			return tool.NewErrorResult(v, name), nil
		}
	}
	return tool.NewErrorResult(tool.NORESULT, name), nil
}

func GetTestResult(toolName, suffix string, fileId bson.ObjectId) (tool.DisplayResult, error) {
	var name string
	switch toolName {
	case junit.NAME:
		f, err := db.File(bson.M{db.ID: fileId}, nil)
		if err != nil {
			return nil, err
		}
		name, _ = util.Extension(f.Name)
	case jacoco.NAME:
		name = jacoco.NAME
	default:
		return nil, fmt.Errorf("unsupported tool %s", toolName)
	}
	name += "-" + suffix
	matcher := bson.M{db.ID: fileId}
	f, err := db.File(matcher, nil)
	if err != nil {
		return nil, err
	}
	ival, ok := f.Results[name]
	if !ok {
		return tool.NewErrorResult(tool.NORESULT, toolName), err
	}
	switch v := ival.(type) {
	case bson.ObjectId:
		//Retrieve result from the db.
		return db.DisplayResult(bson.M{db.ID: v}, nil)
	case string:
		//Error, so create new error result.
		return tool.NewErrorResult(v, toolName), nil
	}
	return tool.NewErrorResult(tool.NORESULT, toolName), nil
}
