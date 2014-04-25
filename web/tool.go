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
	"sort"
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
func tools(pid bson.ObjectId) ([]string, error) {
	p, e := db.Project(bson.M{db.ID: pid}, nil)
	if e != nil {
		return nil, e
	}
	switch tool.Language(p.Lang) {
	case tool.JAVA:
		ts := []string{pmd.NAME, findbugs.NAME, checkstyle.NAME, javac.NAME}
		if _, e := db.JPFConfig(bson.M{db.PROJECTID: pid}, bson.M{db.ID: 1}); e == nil {
			ts = append(ts, jpf.NAME)
		}
		if js, e := db.JUnitTests(bson.M{db.PROJECTID: pid}, bson.M{db.NAME: 1}); e == nil {
			for _, j := range js {
				n, _ := util.Extension(j.Name)
				ts = append(ts, jacoco.NAME+" \u2192 "+n, junit.NAME+" \u2192 "+n)
			}
		}
		sort.Strings(ts)
		return ts, nil
	case tool.C:
		return []string{mk.NAME, gcc.NAME}, nil
	default:
		return nil, fmt.Errorf("unknown language %s", p.Lang)
	}
}

//CreateCheckstyle
func CreateCheckstyle(r *http.Request, c *Context) (string, error) {
	return "", nil
}

//CreateFindbugs
func CreateFindbugs(r *http.Request, c *Context) (string, error) {
	return "", nil
}

func CreateMake(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	_, d, e := ReadFormFile(r, "makefile")
	if e != nil {
		return "Could not read Makefile.", e
	}
	mf := mk.NewMakefile(pid, d)
	if e = db.AddMakefile(mf); e != nil {
		return "Could not create Makefile.", e
	}
	return "Successfully created Makefile.", nil
}

//CreateJUnit adds a new JUnit test for a given project.
func CreateJUnit(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	t, e := getTestType(r)
	if e != nil {
		return e.Error(), e
	}
	n, b, e := ReadFormFile(r, "test")
	if e != nil {
		return "Could not read JUnit file.", e
	}
	//A test does not always need data files.
	var d []byte
	if r.FormValue("data-check") == "true" {
		_, d, e = ReadFormFile(r, "data")
		if e != nil {
			return "Could not read data file.", e
		}
	} else {
		d = make([]byte, 0)
	}
	return "", db.AddJUnitTest(junit.NewTest(pid, n, util.GetPackage(bytes.NewReader(b)), t, b, d))
}

//AddJPF replaces a project's JPF configuration with a provided configuration file.
func AddJPF(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	_, d, e := ReadFormFile(r, "jpf")
	if e != nil {
		return "Could not read JPF configuration file.", e
	}
	return "", db.AddJPFConfig(jpf.NewConfig(pid, d))
}

//CreateJPF replaces a project's JPF configuration with a new, provided configuration.
func CreateJPF(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	//Read JPF properties.
	ps, e := readProperties(r)
	if e != nil {
		return e.Error(), e
	}
	//Convert to JPF property file style.
	d, e := jpf.JPFBytes(ps)
	if e != nil {
		return "Could not create JPF configuration.", e
	}
	if e = db.AddJPFConfig(jpf.NewConfig(pid, d)); e != nil {
		return "Could not create JPF configuration.", e
	}
	return "Successfully created JPF configuration.", nil
}

//readProperties reads JPF properties from a raw string and stores them in a map.
func readProperties(r *http.Request) (map[string][]string, error) {
	p := make(map[string][]string)
	//Read configured listeners and search.
	al, e := GetStrings(r, "addedlisteners")
	if e == nil {
		p["listener"] = al
	}
	s, e := GetString(r, "addedsearches")
	if e == nil {
		p["search.class"] = []string{s}
	}
	o, e := GetString(r, "other")
	if e != nil {
		return p, nil
	}
	ls := strings.Split(o, "\n")
	for _, l := range ls {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		if e = readProperty(l, p); e != nil {
			return nil, e
		}
	}
	return p, nil
}

func readProperty(line string, props map[string][]string) error {
	ps := strings.Split(util.RemoveEmpty(line), "=")
	if len(ps) != 2 {
		return fmt.Errorf("invalid JPF property %s", line)
	}
	k, v := ps[0], ps[1]
	if len(k) == 0 {
		return errors.New("JPF key cannot be empty")
	}
	if len(v) == 0 {
		return fmt.Errorf("JPF value for %s cannot be empty", k)
	}
	if !jpf.Allowed(k) {
		return fmt.Errorf("cannot set JPF property for %s", k)
	}
	sp := strings.Split(v, ",")
	vs := make([]string, 0, len(sp))
	for _, s := range sp {
		if s != "" {
			vs = append(vs, s)
		}
	}
	if v, ok := props[k]; ok {
		props[k] = append(v, vs...)
	} else {
		props[k] = vs
	}
	return nil
}

//CreatePMD creates PMD rules for a project from a provided list.
func CreatePMD(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	s, e := GetStrings(r, "ruleid")
	if e != nil {
		return "Could not read rules.", e
	}
	rs, e := pmd.NewRules(pid, util.ToSet(s))
	if e != nil {
		return "Could not create rules.", e
	}
	if e = db.AddPMDRules(rs); e != nil {
		return "Could not add rules.", e
	}
	return "Successfully added rules.", nil
}

//RunTools runs a tool on submissions in a given project.
//Previous results are deleted if the user has specified that the tool
//should be rerun on all files.
func RunTools(r *http.Request, c *Context) (string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return m, e
	}
	ts, e := GetStrings(r, "tools")
	if e != nil {
		return "Could not read tool.", e
	}
	all, e := tools(pid)
	allTools := e == nil && len(all) == len(ts)
	us, e := GetStrings(r, "users")
	if e != nil {
		return "Could not read tool.", e
	}
	ss, e := db.Submissions(bson.M{db.PROJECTID: pid, db.USER: bson.M{db.IN: us}}, bson.M{db.ID: 1})
	if e != nil {
		return "Could not retrieve submissions.", e
	}
	ct := make([]string, 0, len(ts))
	bt := make([]string, 0, len(ts))
	for _, t := range ts {
		if isChildTool(t, pid) {
			ct = append(ct, t)
		} else {
			bt = append(bt, t)
		}
	}
	redoSubmissions(ss, bt, ct, r.FormValue("runempty-check") != "true", allTools)
	return "Successfully started running tools on submissions.", nil
}

func isChildTool(n string, pid bson.ObjectId) bool {
	return n == jacoco.NAME || db.Contains(db.TESTS, bson.M{db.PROJECTID: pid, db.NAME: n})
}

func redoSubmissions(submissions []*project.Submission, tools, childTools []string, allFiles, allTools bool) {
	for _, s := range submissions {
		var srcs []*project.File
		var e error
		if len(childTools) > 0 {
			srcs, e = db.Files(bson.M{db.SUBID: s.Id, db.TYPE: project.SRC}, bson.M{db.ID: 1}, 0)
			if e != nil {
				util.Log(e)
			}
		}
		fs, e := db.Files(bson.M{db.SUBID: s.Id}, bson.M{db.DATA: 0}, 0)
		if e != nil {
			util.Log(e)
			continue
		}
		for _, f := range fs {
			if allTools {
				runAllTools(f, allFiles)
			} else {
				runTools(f, tools, allFiles)
				if f.Type == project.TEST && len(childTools) > 0 {
					runChildTools(f, srcs, childTools, allFiles)
				}
			}
		}
		if e = processing.RedoSubmission(s.Id); e != nil {
			util.Log(e)
		}
	}
}

func runAllTools(f *project.File, allFiles bool) {
	for r, s := range f.Results {
		rid, isId := s.(bson.ObjectId)
		if !allFiles && isId {
			continue
		}
		delete(f.Results, r)
		if !isId {
			continue
		}
		if e := db.RemoveById(db.RESULTS, rid); e != nil {
			util.Log(e)
		}
	}
	if e := db.Update(db.FILES, bson.M{db.ID: f.Id}, bson.M{db.SET: bson.M{db.RESULTS: f.Results}}); e != nil {
		util.Log(e)
	}
}

func runTools(f *project.File, ts []string, allFiles bool) {
	for _, t := range ts {
		v, ok := f.Results[t]
		if !ok {
			continue
		}
		rid, isId := v.(bson.ObjectId)
		if !allFiles && isId {
			continue
		}
		delete(f.Results, t)
		if !isId {
			continue
		}
		if e := db.RemoveById(db.RESULTS, rid); e != nil {
			util.Log(e)
		}
	}
	if e := db.Update(db.FILES, bson.M{db.ID: f.Id}, bson.M{db.SET: bson.M{db.RESULTS: f.Results}}); e != nil {
		util.Log(e)
	}
}

func runChildTools(test *project.File, srcs []*project.File, childTools []string, allFiles bool) {
	for _, s := range srcs {
		for _, t := range childTools {
			k := t + "-" + s.Id.Hex()
			v, ok := test.Results[k]
			if !ok {
				continue
			}
			rid, isId := v.(bson.ObjectId)
			if !allFiles && isId {
				continue
			}
			delete(test.Results, k)
			if !isId {
				continue
			}
			if e := db.RemoveById(db.RESULTS, rid); e != nil {
				util.Log(e)
			}
		}
	}
	if e := db.Update(db.FILES, bson.M{db.ID: test.Id}, bson.M{db.SET: bson.M{db.RESULTS: test.Results}}); e != nil {
		util.Log(e)
	}
}

//GetResult retrieves a DisplayResult for a given file and result name.
func GetResult(rd *ResultDesc, fileId bson.ObjectId) (tool.DisplayResult, error) {
	m := bson.M{db.ID: fileId}
	f, err := db.File(m, nil)
	if err != nil {
		return nil, err
	}
	s, err := db.Submission(bson.M{db.ID: f.SubId}, nil)
	if err != nil {
		return nil, err
	}
	switch rd.Type {
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
		ival, ok := f.Results[rd.Raw()]
		if !ok {
			return tool.NewErrorResult(tool.NORESULT, rd.Format()), nil
		}
		switch v := ival.(type) {
		case bson.ObjectId:
			//Retrieve result from the db.
			return db.DisplayResult(bson.M{db.ID: v}, nil)
		case string:
			//Error, so create new error result.
			return tool.NewErrorResult(v, rd.Format()), nil
		}
	}
	return tool.NewErrorResult(tool.NORESULT, rd.Format()), nil
}

/*
func getUserTestResult(k string, sid bson.ObjectId) (tool.DisplayResult, error) {
	fs, e := db.Files(bson.M{db.TYPE: project.TEST, db.SUBID: sid, db.RESULTS + "." + k: bson.M{db.EXISTS: true}, db.RESULTS + "." + k: bson.M{db.ISTYPE: 7}}, bson.M{db.DATA: 0}, 1, db.TIME)
	if e != nil || len(fs) == 0 {
		return tool.NewErrorResult(tool.NORESULT, strings.Split(k, "-")[0]), nil
	}
	rid, e := util.GetId(fs[0].Results, k)
	if e != nil {
		return nil, e
	}
	return db.DisplayResult(bson.M{db.ID: rid}, nil)
}

func GetChildResult(key string, fileId bson.ObjectId) (tool.DisplayResult, error) {
	f, e := db.File(bson.M{db.ID: fileId}, nil)
	if e != nil {
		return nil, e
	}
	t := strings.Split(key, ":")[0]
	i, ok := f.Results[key]
	if !ok {
		return tool.NewErrorResult(tool.NORESULT, t), nil
	}
	switch v := i.(type) {
	case bson.ObjectId:
		//Retrieve result from the db.
		return db.DisplayResult(bson.M{db.ID: v}, nil)
	case string:
		//Error, so create new error result.
		return tool.NewErrorResult(v, t), nil
	}
	return tool.NewErrorResult(tool.NORESULT, t), nil
}
*/
func validId(rs ...interface{}) (string, error) {
	for _, i := range rs {
		if r, ok := i.(tool.ToolResult); ok {
			return r.GetId().Hex(), nil
		}
	}
	return "", fmt.Errorf("no valid result found")
}
