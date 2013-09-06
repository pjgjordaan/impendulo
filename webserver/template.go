package webserver

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"html/template"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"strings"
)

var funcs = template.FuncMap{
	"projectName":     projectName,
	"date":            util.Date,
	"setBreaks":       setBreaks,
	"address":         address,
	"base":            filepath.Base,
	"shortname":       shortname,
	"isCode":          isCode,
	"diff":            diff.CalcDiff,
	"createHeader":    fileHeader,
	"sum":             sum,
	"equal":           equal,
	"langs":           tool.Langs,
	"submissionCount": submissionCount,
	"getBusy":         processing.GetStatus,
	"slice":           slice,
	"adjustment":      adjustment,
	"listeners":       jpf.Listeners,
	"searches":        jpf.Searches,
	"rules":           pmd.RuleSet,
	"tools":           tools,
}

const PAGER_SIZE = 10

var (
	templateDir   string
	baseTemplates []string
)

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

func equal(a, b interface{}) bool {
	return a == b
}

func fileHeader(file *project.File) string {
	return file.Name + " " + util.Date(file.Time)
}

func sum(vals ...int) (ret int) {
	for _, val := range vals {
		ret += val
	}
	return
}

func isCode(name string) bool {
	return strings.ToLower(name) == "code"
}

func shortname(exec string) string {
	elements := strings.Split(exec, `.`)
	num := len(elements)
	if num < 2 {
		return exec
	}
	return strings.Join(elements[num-2:], `.`)
}

func address(val interface{}) string {
	return fmt.Sprint(&val)
}

func setBreaks(val string) template.HTML {
	return template.HTML(_setBreaks(val))
}

func _setBreaks(val string) string {
	return strings.Replace(val, "\n", "<br>", -1)
}

func TemplateDir() string {
	if templateDir != "" {
		return templateDir
	}
	templateDir = filepath.Join(StaticDir(), "templates")
	return templateDir
}

func BaseTemplates() []string {
	if baseTemplates != nil {
		return baseTemplates
	}
	baseTemplates = []string{
		filepath.Join(TemplateDir(), "base.html"),
		filepath.Join(TemplateDir(), "index.html"),
		filepath.Join(TemplateDir(), "messages.html"),
	}
	return baseTemplates
}

//T creates a new HTML template from the given files.
func T(names ...string) *template.Template {
	t := template.New("base.html").Funcs(funcs)
	all := make([]string, len(BaseTemplates())+len(names))
	end := copy(all, BaseTemplates())
	for i, name := range names {
		all[i+end] = filepath.Join(TemplateDir(), name+".html")
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
