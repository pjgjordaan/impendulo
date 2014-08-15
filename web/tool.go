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
	"errors"
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
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
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"
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
	JPFKeyError = errors.New("JPF key cannot be empty")
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
				ts = append(ts, jacoco.NAME+":"+n, junit.NAME+":"+n)
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
func CreateCheckstyle(r *http.Request, c *context.C) (string, error) {
	return "", nil
}

//CreateFindbugs
func CreateFindbugs(r *http.Request, c *context.C) (string, error) {
	return "", nil
}

//CreateMake
func CreateMake(r *http.Request, c *context.C) (string, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return "Could not read project id.", e
	}
	_, d, e := webutil.File(r, "makefile")
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
func CreateJUnit(r *http.Request, c *context.C) (string, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return "Could not read project id.", e
	}
	t, e := readTarget(r)
	if e != nil {
		return "Could not read target.", e
	}
	tipe, e := getTestType(r)
	if e != nil {
		return e.Error(), e
	}
	n, b, e := webutil.File(r, "test")
	if e != nil {
		return "Could not read JUnit file.", e
	}
	//A test does not always need data files.
	var d []byte
	if r.FormValue("data-check") == "true" {
		_, d, e = webutil.File(r, "data")
		if e != nil {
			return "Could not read data file.", e
		}
	} else {
		d = make([]byte, 0)
	}
	return "", db.AddJUnitTest(junit.NewTest(pid, n, tipe, t, b, d))
}

//CreateJPF replaces a project's JPF configuration with a new, provided configuration.
func CreateJPF(r *http.Request, c *context.C) (string, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return "Could not read project id.", e
	}
	t, e := readTarget(r)
	if e != nil {
		return "Could not read target.", e
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
	if e = db.AddJPFConfig(jpf.NewConfig(pid, t, d)); e != nil {
		return "Could not create JPF configuration.", e
	}
	return "Successfully created JPF configuration.", nil
}

func readTarget(r *http.Request) (*tool.Target, error) {
	s, e := webutil.String(r, "target")
	if e != nil {
		return nil, e
	}
	pkg, f := util.Extension(s)
	if f == "" {
		if pkg == "" {
			return nil, fmt.Errorf("could not read target from %s", s)
		}
		pkg, f = f, pkg
	}
	return tool.NewTarget(f+".java", pkg, "", tool.JAVA), nil
}

//readProperties reads JPF properties from a raw string and stores them in a map.
func readProperties(r *http.Request) (map[string][]string, error) {
	p := make(map[string][]string)
	//Read configured listeners and search.
	if al, e := webutil.Strings(r, "listeners"); e == nil {
		p["listener"] = al
	}
	if s, e := webutil.String(r, "search"); e == nil {
		p["search.class"] = []string{s}
	}
	o, e := webutil.String(r, "other")
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

//readProperty
func readProperty(line string, props map[string][]string) error {
	ps := strings.Split(util.RemoveEmpty(line), "=")
	if len(ps) != 2 {
		return fmt.Errorf("invalid JPF property %s", line)
	}
	k, v := ps[0], ps[1]
	if len(k) == 0 {
		return JPFKeyError
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
func CreatePMD(r *http.Request, c *context.C) (string, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return "Could not read project id.", e
	}
	s, e := webutil.Strings(r, "rules")
	if e != nil {
		return "Could not read rules.", e
	}
	rs, e := pmd.NewRules(pid, convert.Set(s))
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
func RunTools(r *http.Request, c *context.C) (string, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return "Could not read project id.", e
	}
	ts, e := webutil.Strings(r, "tools")
	if e != nil {
		return "Could not read tool.", e
	}
	all, e := tools(pid)
	allTools := e == nil && len(all) == len(ts)
	us, e := webutil.Strings(r, "users")
	if e != nil {
		return "Could not read users.", e
	}
	ss, e := db.Submissions(bson.M{db.PROJECTID: pid, db.USER: bson.M{db.IN: us}}, bson.M{db.ID: 1})
	if e != nil {
		return "Could not retrieve submissions.", e
	}
	redoSubmissions(ss, ts, r.FormValue("runempty-check") != "true", allTools)
	return "Successfully started running tools on submissions.", nil
}

func addUserTools(sid bson.ObjectId, tools []string) []string {
	ts, e := db.Files(bson.M{db.SUBID: sid, db.TYPE: project.TEST}, bson.M{db.ID: 1, db.NAME: 1}, 0)
	if e != nil {
		return tools
	}
	newTools := make([]string, 0, len(ts)*len(tools))
	for _, tool := range tools {
		if !isUserTool(tool) {
			newTools = append(newTools, tool)
			continue
		}
		n := strings.Split(tool, ":")[1] + ".java"
		for _, t := range ts {
			if n != t.Name {
				continue
			}
			newTools = append(newTools, tool+"-"+t.Id.Hex())
		}
	}
	return newTools
}

func isUserTool(t string) bool {
	if !strings.Contains(t, ":") {
		return false
	}
	return db.Contains(db.TESTS, bson.M{db.NAME: strings.Split(t, ":")[1] + ".java", db.TYPE: junit.USER})
}

//redoSubmissions
func redoSubmissions(submissions []*project.Submission, tools []string, allFiles, allTools bool) {
	for _, s := range submissions {
		var ts []string
		if !allTools {
			ts = addUserTools(s.Id, tools)
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
				runTools(f, ts, allFiles)
			}
		}
		if e = mq.RedoSubmission(s.Id); e != nil {
			util.Log(e)
		}
	}
}

//runAllTools
func runAllTools(f *project.File, allFiles bool) {
	for r, s := range f.Results {
		rid, e := convert.Id(s)
		if !allFiles && e == nil {
			continue
		}
		delete(f.Results, r)
		if e != nil {
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

//runTools
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

//GetResult retrieves a DisplayResult for a given file and result name.
func GetResult(r *context.Result, fileId bson.ObjectId) (result.Displayer, error) {
	m := bson.M{db.ID: fileId}
	f, err := db.File(m, nil)
	if err != nil {
		return nil, err
	}
	s, err := db.Submission(bson.M{db.ID: f.SubId}, nil)
	if err != nil {
		return nil, err
	}
	switch r.Type {
	case result.CODE:
		p, err := db.Project(bson.M{db.ID: s.ProjectId}, nil)
		if err != nil {
			return nil, err
		}
		return result.NewCode(fileId, p.Lang, f.Data), nil
	case diff.NAME:
		return diff.NewResult(f), nil
	default:
		ival, ok := f.Results[r.Raw()]
		if !ok {
			return result.NewError(result.NORESULT, r.Format()), nil
		}
		switch v := ival.(type) {
		case bson.ObjectId:
			//Retrieve result from the db.
			return db.Displayer(bson.M{db.ID: v}, nil)
		case string:
			//Error, so create new error result.
			return result.NewError(v, r.Format()), nil
		}
	}
	return result.NewError(result.NORESULT, r.Format()), nil
}

//validId
func validId(rs ...interface{}) (string, error) {
	for _, i := range rs {
		if r, ok := i.(result.Tooler); ok {
			return r.GetId().Hex(), nil
		}
	}
	return "", fmt.Errorf("no valid result found")
}

func getTestType(r *http.Request) (junit.Type, error) {
	v := strings.ToLower(r.FormValue("testtype"))
	switch v {
	case "default":
		return junit.DEFAULT, nil
	case "admin":
		return junit.ADMIN, nil
	case "user":
		return junit.USER, nil
	default:
		return -1, fmt.Errorf("unsupported test type %s", v)
	}
}
