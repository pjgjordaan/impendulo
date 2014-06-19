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
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"

	"html"
	"html/template"

	"labix.org/v2/mgo/bson"

	"net/url"
	"path/filepath"
	"sort"
	"strings"
)

type (
	view uint
)

var (
	funcs = template.FuncMap{
		"databases":          db.Databases,
		"projectName":        db.ProjectName,
		"date":               util.Date,
		"setBreaks":          func(s string) template.HTML { return template.HTML(setBreaks(s)) },
		"address":            func(i interface{}) string { return fmt.Sprint(&i) },
		"base":               filepath.Base,
		"shortName":          util.ShortName,
		"sum":                sum,
		"percent":            percent,
		"round":              round,
		"langs":              tool.Langs,
		"submissionLang":     submissionLang,
		"sub":                func(id bson.ObjectId) (*project.Submission, error) { return db.Submission(bson.M{db.ID: id}, nil) },
		"submissionCount":    submissionCount,
		"getBusy":            mq.GetStatus,
		"slice":              slice,
		"adjustment":         adjustment,
		"tools":              tools,
		"configtools":        configTools,
		"unescape":           html.UnescapeString,
		"escape":             url.QueryEscape,
		"snapshots":          func(id bson.ObjectId) (int, error) { return db.FileCount(id, project.SRC) },
		"launches":           func(id bson.ObjectId) (int, error) { return db.FileCount(id, project.LAUNCH) },
		"usertests":          func(id bson.ObjectId) (int, error) { return db.FileCount(id, project.TEST) },
		"html":               func(s string) template.HTML { return template.HTML(s) },
		"string":             func(b []byte) string { return string(bytes.TrimSpace(b)) },
		"args":               args,
		"insert":             insert,
		"isError":            isError,
		"hasChart":           tool.HasChart,
		"projectSubmissions": projectSubmissions,
		"userSubmissions":    userSubmissions,
		"fileinfos":          _fileinfos,
		"projects":           projects,
		"langProjects": func(l string) ([]*project.Project, error) {
			return db.Projects(bson.M{db.LANG: l}, nil, db.NAME)
		},
		"resultNames":   db.ResultNames,
		"users":         func() ([]string, error) { return db.Usernames(nil) },
		"skeletons":     skeletons,
		"overviewChart": overviewChart,
		"typeCounts":    db.TypeCounts,
		"editables":     func() []string { return []string{"Project", "User"} },
		"permissions":   user.Permissions,
		"file":          func(id bson.ObjectId) (*project.File, error) { return db.File(bson.M{db.ID: id}, nil) },
		"toTitle":       util.Title,
		"addSpaces":     func(s string) string { return strings.Join(util.SplitTitles(s), " ") },
		"chartTime":     chartTime,
		"validId":       validId,
		"emptyM": func(m map[string][]interface{}) bool {
			return len(m) == 0
		},
		"emptyS": func(s []interface{}) bool {
			return len(s) == 0
		},
		"empty":     func(s string) bool { return strings.TrimSpace(s) == "" },
		"sortFiles": sortFiles,
		"project":   func(id bson.ObjectId) (*project.Project, error) { return db.Project(bson.M{db.ID: id}, nil) },
	}
	templateDir      string
	baseTemplates    []string
	InvalidArgsError = errors.New("invalid args call")
)

func _fileinfos(sid bson.ObjectId) ([]*db.FileInfo, error) {
	return db.FileInfos(bson.M{db.SUBID: sid, db.TYPE: bson.M{db.IN: []project.Type{project.SRC, project.TEST}}})
}

func pagerSize(high int) int {
	if high < 100 {
		return 10
	} else if high < 1000 {
		return 8
	}
	return 6
}

//slice gets a subslice from files which is at most PAGER_SIZE
//in length with selected within in the subslice.
func slice(files []*project.File, i int) []*project.File {
	s := pagerSize(i)
	if len(files) < s {
		return files
	} else if i < s/2 {
		return files[:s]
	} else if i+s/2 >= len(files) {
		return files[len(files)-s:]
	}
	return files[i-s/2 : i+s/2]
}

//adjustment determines the adjustment slice will make to files.
func adjustment(files []*project.File, i int) int {
	s := pagerSize(i)
	if len(files) < s || i < s/2 {
		return 0
	} else if i+s/2 >= len(files) {
		return len(files) - s
	}
	return i - s/2
}

func sortFiles(ids []interface{}) []*project.File {
	fs := make([]*project.File, 0, len(ids))
	for _, s := range ids {
		id, e := convert.Id(s)
		if e != nil {
			continue
		}
		f, e := db.File(bson.M{db.ID: id}, bson.M{db.TIME: 1})
		if e != nil {
			continue
		}
		fs = append(fs, f)
	}
	sort.Sort(files(fs))
	return fs
}

type files []*project.File

func (f files) Less(i, j int) bool {
	return f[i].Time >= f[j].Time
}

func (f files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f files) Len() int {
	return len(f)
}

func chartTime(f *project.File) (float64, error) {
	s, e := db.Submission(bson.M{db.ID: f.SubId}, nil)
	if e != nil {
		return -1.0, e
	}
	return util.Round(float64(f.Time-s.Time)/1000.0, 2), nil
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

func userSubmissions(uid string) ([]*project.Submission, error) {
	return db.Submissions(bson.M{db.USER: uid}, nil, "-"+db.TIME)
}

func projectSubmissions(pid bson.ObjectId) ([]*project.Submission, error) {
	return db.Submissions(bson.M{db.PROJECTID: pid}, nil, "-"+db.TIME)
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
		return nil, InvalidArgsError
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

//submissionCount
func submissionCount(id interface{}) (int, error) {
	switch t := id.(type) {
	case bson.ObjectId:
		return db.Count(db.SUBMISSIONS, bson.M{db.PROJECTID: t})
	case string:
		return db.Count(db.SUBMISSIONS, bson.M{db.USER: t})
	default:
		return -1, fmt.Errorf("Unknown id type %q.", id)
	}
}

//sum calculates the sum of vals.
func sum(vals ...interface{}) (int, error) {
	s := 0
	for _, i := range vals {
		v, e := convert.Int(i)
		if e != nil {
			return 0, e
		}
		s += v
	}
	return s, nil
}

func percent(a, b interface{}) (float64, error) {
	c, e := convert.Float64(a)
	if e != nil {
		return 0.0, e
	}
	d, e := convert.Float64(b)
	if e != nil {
		return 0.0, e
	}
	return d / c * 100.0, nil
}

func round(i, ip interface{}) (float64, error) {
	x, e := convert.Float64(i)
	if e != nil {
		return 0.0, e
	}
	p, e := convert.Int(ip)
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
