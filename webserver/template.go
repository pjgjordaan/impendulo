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
	"errors"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"html"
	"html/template"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"strings"
)

type (
	//Args represents arguments passed to html templates.
	Args map[string]interface{}
)

var (
	funcs = template.FuncMap{
		"projectName":           projectName,
		"date":                  util.Date,
		"setBreaks":             func(s string) template.HTML { return template.HTML(setBreaks(s)) },
		"address":               func(i interface{}) string { return fmt.Sprint(&i) },
		"base":                  filepath.Base,
		"shortName":             util.ShortName,
		"sum":                   sum,
		"langs":                 tool.Langs,
		"submissionCount":       submissionCount,
		"getBusy":               processing.GetStatus,
		"slice":                 slice,
		"adjustment":            adjustment,
		"listeners":             jpf.Listeners,
		"searches":              jpf.Searches,
		"rules":                 pmd.RuleSet,
		"tools":                 tools,
		"unescape":              html.UnescapeString,
		"snapshots":             func(id bson.ObjectId) (int, error) { return fileCount(id, project.SRC) },
		"launches":              func(id bson.ObjectId) (int, error) { return fileCount(id, project.LAUNCH) },
		"html":                  func(s string) template.HTML { return template.HTML(s) },
		"string":                func(b []byte) string { return string(b) },
		"args":                  args,
		"insert":                insert,
		"isError":               isError,
		"hasChart":              func(n string) bool { return n != tool.CODE && n != diff.NAME && n != tool.SUMMARY },
		"submissions":           projectSubmissions,
		"projects":              func() ([]*project.Project, error) { return db.Projects(nil, bson.M{project.SKELETON: 0}, project.NAME) },
		"users":                 func() ([]*user.User, error) { return db.Users(nil, user.ID) },
		"displayResult":         displayResult,
		"displayCodeBug":        displayCodeBug,
		"getFiles":              func(subId bson.ObjectId) string { return fmt.Sprintf("getfiles?sid=%s", subId.Hex()) },
		"getUserChart":          func(user string) string { return fmt.Sprintf("getsubmissionschart?uid=%s", user) },
		"getProjectChart":       func(id bson.ObjectId) string { return fmt.Sprintf("getsubmissionschart?pid=%s", id.Hex()) },
		"getUserSubmissions":    func(user string) string { return fmt.Sprintf("getsubmissions?uid=%s", user) },
		"getProjectSubmissions": func(id bson.ObjectId) string { return fmt.Sprintf("getsubmissions?pid=%s", id.Hex()) },
		"singleChart":           singleChart,
		"compareChart":          compareChart,
		"collections":           db.Collections,
		"overviewChart":         overviewChart,
		"typeCounts":            TypeCounts,
	}
	templateDir   string
	baseTemplates []string
)

const (
	PAGER_SIZE = 10
)

func compareChart(sid bson.ObjectId, uid, result, file, compare string) string {
	return singleChart(sid, uid, result, file) + fmt.Sprintf("&compare=%s", compare)
}

func singleChart(sid bson.ObjectId, uid, result, file string) string {
	return fmt.Sprintf("displaychart?sid=%s&uid=%s&result=%s&file=%s", sid.Hex(), uid, result, file)
}

func displayResult(sid bson.ObjectId, uid, result, file string, current, next int) string {
	return fmt.Sprintf("displayresult?sid=%s&uid=%s&result=%s&file=%s&current=%d&next=%d",
		sid.Hex(), uid, result, file, current, next)
}

func displayCodeBug(sid bson.ObjectId, uid, result, file string, current, next int, resultId, bugId string, index int) string {
	return displayResult(sid, uid, result, file, current, next) +
		fmt.Sprintf("&rid=%s&bid=%s&bindex=%d", resultId, bugId, index)
}

func projectSubmissions(id bson.ObjectId) (subs []*project.Submission, err error) {
	matcher := bson.M{project.PROJECT_ID: id}
	subs, err = db.Submissions(matcher, nil, "-"+project.TIME)
	return
}

//isError checks whether a result is an ErrorResult.
func isError(result tool.DisplayResult) bool {
	_, ok := result.(*tool.ErrorResult)
	return ok
}

//args creates a map from the list of items. Items at even indices in the list
//must be strings and are keys in the map while the item which immediately follows them
//will be the value which corresponds to that key in the map. The list must therefore
//contain an even number of items.
func args(values ...interface{}) (ret Args, err error) {
	if len(values)%2 != 0 {
		err = errors.New("Invalid args call.")
		return
	}
	ret = make(Args, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			err = errors.New("args keys must be strings.")
			return
		}
		ret[key] = values[i+1]
	}
	return
}

//insert adds a key-value pair to the specified map.
func insert(args Args, key string, value interface{}) Args {
	args[key] = value
	return args
}

//fileCount
func fileCount(subId bson.ObjectId, tipe project.Type) (int, error) {
	return db.Count(
		db.FILES,
		bson.M{
			project.SUBID: subId,
			project.TYPE:  tipe,
		},
	)
}

//slice gets a subslice from files which is at most PAGER_SIZE
//in length with selected within in the subslice.
func slice(files []*project.File, selected int) (ret []*project.File) {
	if len(files) < PAGER_SIZE {
		ret = files
	} else if selected < PAGER_SIZE/2 {
		ret = files[:PAGER_SIZE]
	} else if selected+PAGER_SIZE/2 >= len(files) {
		ret = files[len(files)-PAGER_SIZE:]
	} else {
		ret = files[selected-PAGER_SIZE/2 : selected+PAGER_SIZE/2]
	}
	return
}

//adjustment determines the adjustment slice will make to files.
func adjustment(files []*project.File, selected int) (ret int) {
	if len(files) < PAGER_SIZE || selected < PAGER_SIZE/2 {
		ret = 0
	} else if selected+PAGER_SIZE/2 >= len(files) {
		ret = len(files) - PAGER_SIZE
	} else {
		ret = selected - PAGER_SIZE/2
	}
	return
}

//submissionCount
func submissionCount(id interface{}) (int, error) {
	switch tipe := id.(type) {
	case bson.ObjectId:
		return db.Count(db.SUBMISSIONS, bson.M{project.PROJECT_ID: tipe})
	case string:
		return db.Count(db.SUBMISSIONS, bson.M{project.USER: tipe})
	default:
		return -1, fmt.Errorf("Unknown id type %q.", id)
	}
}

//sum calculates the sum of vals.
func sum(vals ...int) (ret int) {
	for _, val := range vals {
		ret += val
	}
	return
}

//setBreaks replaces newlines with HTML break tags.
func setBreaks(val string) string {
	return strings.Replace(val, "\n", "<br>", -1)
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
	baseTemplates = []string{
		filepath.Join(TemplateDir(), "base.html"),
		filepath.Join(TemplateDir(), "index.html"),
		filepath.Join(TemplateDir(), "messages.html"),
		filepath.Join(TemplateDir(), "footer.html"),
		filepath.Join(TemplateDir(), "breadcrumb.html"),
	}
	return baseTemplates
}

//T creates a new HTML template from the given files.
func T(names ...string) *template.Template {
	t := template.New("base.html").Funcs(funcs)
	all := make([]string, len(BaseTemplates()), len(BaseTemplates())+len(names))
	copy(all, BaseTemplates())
	for _, name := range names {
		path := filepath.Join(TemplateDir(), name+".html")
		if util.Exists(path) {
			all = append(all, path)
		}
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
