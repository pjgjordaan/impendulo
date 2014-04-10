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
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"

	"html"
	"html/template"

	"labix.org/v2/mgo/bson"

	"net/url"
	"path/filepath"
	"strings"
)

type (
	view uint
)

var (
	funcs = template.FuncMap{
		"projectName":     projectName,
		"date":            util.Date,
		"setBreaks":       func(s string) template.HTML { return template.HTML(setBreaks(s)) },
		"address":         func(i interface{}) string { return fmt.Sprint(&i) },
		"base":            filepath.Base,
		"shortName":       util.ShortName,
		"sum":             sum,
		"percent":         percent,
		"round":           round,
		"langs":           tool.Langs,
		"submissionLang":  submissionLang,
		"submissionCount": submissionCount,
		"getBusy":         processing.GetStatus,
		"slice":           slice,
		"adjustment":      adjustment,
		"listeners":       jpf.Listeners,
		"searches":        jpf.Searches,
		"rules":           pmd.RuleSet,
		"tools":           tools,
		"configtools":     configTools,
		"unescape":        html.UnescapeString,
		"escape":          url.QueryEscape,
		"snapshots":       func(id bson.ObjectId) (int, error) { return fileCount(id, project.SRC) },
		"launches":        func(id bson.ObjectId) (int, error) { return fileCount(id, project.LAUNCH) },
		"html":            func(s string) template.HTML { return template.HTML(s) },
		"string":          func(b []byte) string { return string(bytes.TrimSpace(b)) },
		"args":            args,
		"insert":          insert,
		"isError":         isError,
		"hasChart":        hasChart,
		"submissions":     projectSubmissions,
		"files":           submissionFiles,
		"projects":        projects,
		"langProjects": func(l string) ([]*project.Project, error) {
			return db.Projects(bson.M{db.LANG: l}, nil, db.NAME)
		},
		"analysisNames":         analysisNames,
		"users":                 func() ([]*user.User, error) { return db.Users(nil, user.ID) },
		"skeletons":             skeletons,
		"getFiles":              func(sid bson.ObjectId) string { return fmt.Sprintf("getfiles?subid=%s", sid.Hex()) },
		"getUserChart":          func(u string) string { return fmt.Sprintf("getsubmissionschart?userid=%s", u) },
		"getProjectChart":       func(pid bson.ObjectId) string { return fmt.Sprintf("getsubmissionschart?projectid=%s", pid.Hex()) },
		"getUserSubmissions":    func(u string) string { return fmt.Sprintf("getsubmissions?userid=%s", u) },
		"getProjectSubmissions": func(pid bson.ObjectId) string { return fmt.Sprintf("getsubmissions?projectid=%s", pid.Hex()) },
		"collections":           db.Collections,
		"overviewChart":         overviewChart,
		"typeCounts":            TypeCounts,
		"editables":             func() []string { return []string{"Project", "User", "Submission", "File"} },
		"permissions":           user.Permissions,
		"file":                  func(id bson.ObjectId) (*project.File, error) { return db.File(bson.M{db.ID: id}, nil) },
		"toTitle":               util.Title,
		"addSpaces":             func(s string) string { return strings.Join(util.SplitTitles(s), " ") },
		"chartTime":             chartTime,
	}
	templateDir   string
	baseTemplates []string
)

const (
	PAGER_SIZE      = 10
	testView   view = iota
	analysisView
)

func chartTime(f *project.File) (float64, error) {
	s, e := db.Submission(bson.M{db.ID: f.SubId}, nil)
	if e != nil {
		return -1.0, e
	}
	return util.Round(float64(f.Time-s.Time)/1000.0, 2), nil
}

func analysisNames(pid bson.ObjectId, t project.Type) ([]string, error) {
	switch t {
	case project.TEST:
		return resultNames(pid, testView)
	case project.SRC:
		return resultNames(pid, analysisView)
	}
	return nil, fmt.Errorf("unsupported file type %s", t)
}

func resultNames(pid bson.ObjectId, v view) ([]string, error) {
	switch v {
	case testView:
		return []string{javac.NAME, tool.CODE, diff.NAME, junit.NAME, jacoco.NAME}, nil
	case analysisView:
		return db.AllResultNames(pid)
	}
	return nil, fmt.Errorf("unsupported view type %d", v)
}

func submissionLang(sid bson.ObjectId) (string, error) {
	s, e := db.Submission(bson.M{db.ID: sid}, nil)
	if e != nil {
		return "", e
	}
	p, e := db.Project(bson.M{db.ID: s.ProjectId}, nil)
	if e != nil {
		return "", e
	}
	return strings.ToLower(p.Lang), nil
}

func projects() ([]*project.Project, error) {
	return db.Projects(nil, nil, db.NAME)
}

func skeletons(pid bson.ObjectId) ([]*project.Skeleton, error) {
	return db.Skeletons(bson.M{db.PROJECTID: pid}, bson.M{db.DATA: 0}, db.NAME)
}

func projectSubmissions(pid bson.ObjectId) ([]*project.Submission, error) {
	return db.Submissions(bson.M{db.PROJECTID: pid}, nil, "-"+db.TIME)
}

func submissionFiles(sid bson.ObjectId) ([]*project.File, error) {
	return db.Files(bson.M{db.SUBID: sid}, nil, "-"+db.TIME)
}

//isError checks whether a result is an ErrorResult.
func isError(i interface{}) bool {
	_, ok := i.(*tool.ErrorResult)
	return ok
}

//args creates a map from the list of items. Items at even indices in the list
//must be strings and are keys in the map while the item which immediately follows them
//will be the value which corresponds to that key in the map. The list must therefore
//contain an even number of items.
func args(values ...interface{}) (Args, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("Invalid args call.")
	}
	a := make(Args, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		if k, ok := values[i].(string); ok {
			a[k] = values[i+1]
		} else {
			return nil, fmt.Errorf("key %v is not a string", values[i])
		}
	}
	return a, nil
}

//insert adds a key-value pair to the specified map.
func insert(a Args, k string, v interface{}) Args {
	a[k] = v
	return a
}

//fileCount
func fileCount(sid bson.ObjectId, t project.Type) (int, error) {
	return db.Count(db.FILES, bson.M{db.SUBID: sid, db.TYPE: t})
}

//slice gets a subslice from files which is at most PAGER_SIZE
//in length with selected within in the subslice.
func slice(files []*project.File, i int) []*project.File {
	if len(files) < PAGER_SIZE {
		return files
	} else if i < PAGER_SIZE/2 {
		return files[:PAGER_SIZE]
	} else if i+PAGER_SIZE/2 >= len(files) {
		return files[len(files)-PAGER_SIZE:]
	}
	return files[i-PAGER_SIZE/2 : i+PAGER_SIZE/2]
}

//adjustment determines the adjustment slice will make to files.
func adjustment(files []*project.File, i int) int {
	if len(files) < PAGER_SIZE || i < PAGER_SIZE/2 {
		return 0
	} else if i+PAGER_SIZE/2 >= len(files) {
		return len(files) - PAGER_SIZE
	}
	return i - PAGER_SIZE/2
}

//submissionCount
func submissionCount(id interface{}) (int, error) {
	switch tipe := id.(type) {
	case bson.ObjectId:
		return db.Count(db.SUBMISSIONS, bson.M{db.PROJECTID: tipe})
	case string:
		return db.Count(db.SUBMISSIONS, bson.M{db.USER: tipe})
	}
	return -1, fmt.Errorf("Unknown id type %q.", id)
}

//sum calculates the sum of vals.
func sum(vals ...interface{}) (int, error) {
	s := 0
	for _, i := range vals {
		v, e := util.Int(i)
		if e != nil {
			return 0, e
		}
		s += v
	}
	return s, nil
}

func percent(a, b interface{}) (float64, error) {
	c, e := util.Float64(a)
	if e != nil {
		return 0.0, e
	}
	d, e := util.Float64(b)
	if e != nil {
		return 0.0, e
	}
	return d / c * 100.0, nil
}

func round(i, ip interface{}) (float64, error) {
	x, e := util.Float64(i)
	if e != nil {
		return 0.0, e
	}
	p, e := util.Int(ip)
	if e != nil {
		return 0.0, e
	}
	return util.Round(x, p), nil
}

//setBreaks replaces newlines with HTML break tags.
func setBreaks(s string) string {
	return strings.Replace(s, "\n", "<br>", -1)
}

//TemplateDir retrieves the directory
//in which HTML templates are stored.
func TemplateDir() string {
	if templateDir != "" {
		return templateDir
	}
	templateDir = filepath.Join(StaticDir(), "html")
	return templateDir
}

//BaseTemplates loads the base HTML templates used by all
//views in the web app.
func BaseTemplates() []string {
	if baseTemplates != nil {
		return baseTemplates
	}
	t := TemplateDir()
	baseTemplates = []string{
		filepath.Join(t, "base.html"),
		filepath.Join(t, "index.html"),
		filepath.Join(t, "messages.html"),
		filepath.Join(t, "footer.html"),
		filepath.Join(t, "breadcrumb.html"),
	}
	return baseTemplates
}

//T creates a new HTML template from the given files.
func T(names ...string) *template.Template {
	t := template.New("base.html").Funcs(funcs)
	all := make([]string, len(BaseTemplates()), len(BaseTemplates())+len(names))
	copy(all, BaseTemplates())
	for _, n := range names {
		p := filepath.Join(TemplateDir(), n+".html")
		if util.Exists(p) {
			all = append(all, p)
		}
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
