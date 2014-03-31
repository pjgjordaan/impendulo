package web

import (
	"code.google.com/p/gorilla/pat"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	AJAX func(*http.Request) ([]byte, error)
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
		"chart": getChart, "tools": getTools, "users": getUsers,
		"skeletons": getSkeletons, "code": getCode, "submissions": submissions,
	}
	for n, f := range fs {
		r.Add("GET", "/"+n, f)
	}
}

func submissions(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	s, e := db.Submissions(bson.M{db.PROJECTID: pid}, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"submissions": s})
}

func getUsers(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	u, e := users(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

func getTools(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	t, e := tools(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tools": t})
}

func getCode(r *http.Request) ([]byte, error) {
	rid, _, e := getId(r, "resultid", "result")
	if e != nil {
		return nil, e
	}
	tr, e := db.ToolResult(bson.M{db.ID: rid}, bson.M{db.FILEID: 1})
	if e != nil {
		return nil, e
	}
	f, e := db.File(bson.M{db.ID: tr.GetFileId()}, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(f.Data)})
}

func getSkeletons(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
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
	e := r.ParseForm()
	if e != nil {
		return nil, e
	}
	sid, _, e := getSubId(r)
	if e != nil {
		return nil, e
	}
	n, e := GetString(r, "file")
	if e != nil {
		return nil, e
	}
	rn, e := GetString(r, "result")
	if e != nil {
		return nil, e
	}
	switch rn {
	case jacoco.NAME:
		cId, e := util.ReadId(r.FormValue("childfileid"))
		if e != nil {
			return nil, e
		}
		rn += "-" + cId.Hex()
	case junit.NAME:
		rn, _ = util.Extension(n)
		if cId, e := util.ReadId(r.FormValue("childfileid")); e == nil {
			rn += "-" + cId.Hex()
		}
	}
	fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: n}, bson.M{db.DATA: 0}, db.TIME)
	if e != nil {
		return nil, e
	}
	c, e := LoadChart(rn, fs)
	if e != nil {
		return nil, e
	}
	if subs, ok := r.Form["submissions[]"]; ok {
		for _, s := range subs {
			sid, e := util.ReadId(s)
			if e != nil {
				return nil, e
			}
			fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: n}, bson.M{db.DATA: 0})
			if e != nil {
				return nil, e
			}
			nc, e := LoadChart(rn, fs)
			if e != nil {
				return nil, e
			}
			c = append(c, nc...)
		}
	}
	m := map[string]interface{}{"chart": c}
	return util.JSON(m)
}
