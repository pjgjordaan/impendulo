package ajax

import (
	"code.google.com/p/gorilla/pat"

	"errors"
	"strings"

	"fmt"

	"github.com/godfried/impendulo/util"

	"net/http"
)

type (
	Get    func(*http.Request) ([]byte, error)
	Post   func(http.ResponseWriter, *http.Request) error
	Select struct {
		Id, Name string
		User     bool
	}
	Selects []*Select
)

const (
	LOG_AJAX = "webserver/ajax/ajax.go"
)

var (
	CountsError   = errors.New("unsupported counts request")
	CommentsError = errors.New("unsupported comments request")
	ResultsError  = errors.New("cannot retrieve results")
)

func (s Selects) Len() int {
	return len(s)
}

func (s Selects) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Selects) Less(i, j int) bool {
	return strings.ToLower(s[i].Name) <= strings.ToLower(s[j].Name)
}

func (a Get) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, e := a(r)
	if e != nil {
		util.Log(e, LOG_AJAX)
		b, _ = util.JSON(map[string]interface{}{"error": e.Error()})
	}
	fmt.Fprint(w, string(b))
}

func (a Post) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := a(w, r); e != nil {
		b, _ := util.JSON(map[string]interface{}{"error": e.Error()})
		util.Log(e, LOG_AJAX)
		fmt.Fprint(w, string(b))
	}
}

func Generate(r *pat.Router) {
	gets := map[string]Get{
		"chart-data": Chart, "usernames": Usernames, "collections": Collections, "pmdrules": PMDRules,
		"skeletons": Skeletons, "submissions": Submissions, "results": Results, "jpflisteners": JPFListeners,
		"langs": Langs, "projects": Projects, "files": Files, "tools": Tools, "jpfsearches": JPFSearches,
		"code": Code, "users": Users, "permissions": Perms, "comparables": Comparables,
		"tests": Tests, "test-types": TestTypes, "filenames": FileNames, "basicfileinfos": BasicFileInfos, "status": Status, "counts": Counts,
		"comments": Comments, "fileresults": FileResults, "chart-options": ChartOptions, "assignments": Assignments,
		"typecounts": TypeCounts, "fileinfos": FileInfos, "resultnames": ResultNames, "databases": Databases,
	}
	for n, f := range gets {
		r.Add("GET", "/"+n, f)
	}
	posts := map[string]Post{
		"setcontext": SetContext, "addcomment": Comment,
	}
	for n, f := range posts {
		r.Add("POST", "/"+n, f)
	}
}
