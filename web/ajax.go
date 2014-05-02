package web

import (
	"code.google.com/p/gorilla/pat"

	"errors"

	"strings"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	AJAX   func(*http.Request) ([]byte, error)
	Select struct {
		Id, Name string
	}
)

func (a AJAX) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, e := a(r)
	if e != nil {
		util.Log(e, LOG_HANDLERS)
		b, _ = util.JSON(map[string]interface{}{"error": e.Error()})
	}
	fmt.Fprint(w, string(b))
}

func GenerateAJAX(r *pat.Router) {
	fs := map[string]AJAX{
		"chart": getChart, "usernames": getUsernames, "collections": collections,
		"skeletons": getSkeletons, "submissions": submissions, "results": ajaxResults,
		"langs": getLangs, "projects": getProjects, "files": ajaxFiles, "tools": ajaxTools,
		"code": ajaxCode, "users": ajaxUsers, "permissions": ajaxPerms, "comparables": ajaxComparables,
	}
	for n, f := range fs {
		r.Add("GET", "/"+n, f)
	}
}

func ajaxResults(r *http.Request) ([]byte, error) {
	var rs []string
	if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
		rs = db.ProjectResults(pid)
	} else if u, e := GetString(r, "user-id"); e == nil {
		rs = db.UserResults(u)
	} else {
		return nil, errors.New("cannot retrieve results")
	}
	s := make([]*Select, len(rs))
	for i, r := range rs {
		s[i] = &Select{Id: r, Name: strings.Replace(r, ":", "\u2192", -1)}
	}
	return util.JSON(map[string]interface{}{"results": s})
}

func ajaxComparables(r *http.Request) ([]byte, error) {
	id, e := convert.Id(r.FormValue("id"))
	if e != nil {
		return nil, e
	}
	tr, e := db.ToolResult(bson.M{db.ID: id}, nil)
	if e != nil {
		return nil, e
	}
	if tr.GetType() != jacoco.NAME && tr.GetType() != junit.NAME {
		return util.JSON(map[string]interface{}{"comparables": []*Select{}})
	}
	f, e := db.File(bson.M{db.ID: tr.GetFileId()}, bson.M{db.SUBID: 1})
	if e != nil {
		return nil, e
	}
	s, e := db.Submission(bson.M{db.ID: f.SubId}, bson.M{db.PROJECTID: 1})
	if e != nil {
		return nil, e
	}
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: s.ProjectId}, bson.M{db.NAME: 1, db.TYPE: 1})
	if e != nil {
		return nil, e
	}
	cmp := make([]*Select, 0, len(ts))
	for _, t := range ts {
		n, _ := util.Extension(t.Name)
		if n == tr.GetName() {
			continue
		}
		cmp = append(cmp, &Select{tr.GetType() + ":" + n, tr.GetType() + " \u2192 " + n})
	}
	return util.JSON(map[string]interface{}{"comparables": cmp})
}

func ajaxPerms(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"permissions": user.PermissionInfos()})
}

func ajaxUsers(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if n, e := GetString(r, "name"); e == nil {
		m[db.ID] = n
	}
	u, e := db.Users(m)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

func ajaxCode(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if rid, e := convert.Id(r.FormValue("result-id")); e == nil {
		tr, e := db.ToolResult(bson.M{db.ID: rid}, bson.M{db.FILEID: 1})
		if e != nil {
			return nil, e
		}
		m[db.ID] = tr.GetFileId()
	}
	if id, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = id
	}
	if n, e := GetString(r, "tool-name"); e == nil {
		d, e := NewResultDesc(n)
		if e != nil {
			return nil, e
		}
		if d.FileID != "" {
			m[db.ID] = d.FileID
		} else if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
			return loadTestCode(pid, d.Name)
		} else {
			return nil, fmt.Errorf("could not load code for %s", d.Format())
		}
	}
	f, e := db.File(m, bson.M{db.DATA: 1})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(f.Data)})
}

func loadTestCode(pid bson.ObjectId, n string) ([]byte, error) {
	t, e := db.JUnitTest(bson.M{db.PROJECTID: pid, db.NAME: n + ".java"}, bson.M{db.TEST: 1})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(t.Test)})
}

func collections(r *http.Request) ([]byte, error) {
	n, e := GetString(r, "db")
	if e != nil {
		return nil, e
	}
	c, e := db.Collections(n)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"collections": c})
}

func getProjects(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = pid
	}
	p, e := db.Projects(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"projects": p})
}

func ajaxTools(r *http.Request) ([]byte, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return nil, e
	}
	t, e := tools(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tools": t})
}

func ajaxFiles(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if sid, e := convert.Id(r.FormValue("submission-id")); e == nil {
		m[db.SUBID] = sid
	}
	if id, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = id
	}
	format, _ := GetString(r, "format")
	var f interface{}
	var e error
	if format == "nested" {
		f, e = nestedFiles(m)
	} else {
		f, e = db.Files(m, bson.M{db.DATA: 0}, 0, db.TIME)
	}
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"files": f})
}

func nestedFiles(m bson.M) (map[project.Type]map[string][]*project.File, error) {
	ts := []project.Type{project.SRC, project.LAUNCH, project.ARCHIVE, project.TEST}
	ns, e := db.FileNames(m)
	if e != nil {
		return nil, e
	}
	fm := make(map[project.Type]map[string][]*project.File)
	for _, t := range ts {
		nm := make(map[string][]*project.File)
		m[db.TYPE] = t
		for _, n := range ns {
			m[db.NAME] = n
			fs, e := db.Files(m, bson.M{db.DATA: 0}, 0, db.TIME)
			if e == nil && len(fs) > 0 {
				nm[n] = fs
			}
		}
		if len(nm) > 0 {
			fm[t] = nm
		}
	}
	return fm, nil
}

func submissions(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if sid, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = sid
	}
	if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
		m[db.PROJECTID] = pid
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"submissions": s})
}

func getUsernames(r *http.Request) ([]byte, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	var u []string
	if e != nil {
		u, e = db.Usernames(nil)
	} else {
		u, e = projectUsernames(pid)
	}
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"usernames": u})
}

func getLangs(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"langs": []tool.Language{tool.JAVA, tool.C}})
}

func getSkeletons(r *http.Request) ([]byte, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return nil, e
	}
	s, e := skeletons(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"skeletons": s})
}

func getChart(r *http.Request) ([]byte, error) {
	if e := r.ParseForm(); e != nil {
		return nil, e
	}
	t, e := GetString(r, "type")
	if e != nil {
		return nil, e
	}
	switch t {
	case "file":
		return fileChart(r)
	case "submission":
		return submissionChart(r)
	default:
		return nil, fmt.Errorf("unsupported chart type %s", t)
	}
}

func submissionChart(r *http.Request) ([]byte, error) {
	rn, e := GetString(r, "result")
	if e != nil {
		return nil, e
	}
	rd, e := NewResultDesc(rn)
	if e != nil {
		return nil, e
	}
	m := bson.M{}
	t, e := GetString(r, "submission-type")
	if e != nil {
		return nil, e
	}
	id, e := GetString(r, "id")
	if e != nil {
		return nil, e
	}
	switch t {
	case "project":
		pid, e := convert.Id(id)
		if e != nil {
			return nil, e
		}
		m[db.PROJECTID] = pid
	case "user":
		m[db.USER] = id
	default:
		return nil, errors.New("no submission chart type specified")
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	c, e := SubmissionChart(s, rd)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"chart": c})
}

func fileChart(r *http.Request) ([]byte, error) {
	fn, e := GetString(r, "file")
	if e != nil {
		return nil, e
	}
	rn, e := GetString(r, "result")
	if e != nil {
		return nil, e
	}
	subs, e := GetStrings(r, "submissions[]")
	if e != nil {
		return nil, e
	}
	var d ChartData
	var first bson.ObjectId
	for i, s := range subs {
		r := rn
		m := bson.M{db.NAME: fn}
		id, e := convert.Id(s)
		if e != nil {
			id = first
			if !strings.Contains(s, ":") {
				m[db.NAME] = s
			} else {
				r = s
			}
		}
		if i == 0 {
			first = id
		}
		m[db.SUBID] = id
		fs, e := db.Files(m, bson.M{db.DATA: 0}, 0, db.TIME)
		if e != nil {
			return nil, e
		}
		c, e := LoadChart(r, fs)
		if e != nil {
			return nil, e
		}
		d = append(d, c...)
	}
	return util.JSON(map[string]interface{}{"chart": d})
}
