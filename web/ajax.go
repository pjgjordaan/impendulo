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

func (a AJAX) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b, e := a(req)
	if e != nil {
		util.Log(e, LOG_HANDLERS)
		b, _ = util.JSON(map[string]interface{}{"error": e.Error()})
	}
	fmt.Fprint(w, string(b))
}

func GenerateAJAX(router *pat.Router) {
	ajaxFuncs := map[string]AJAX{
		"chart": getChart, "tools": getTools, "users": getUsers,
		"skeletons": getSkeletons, "code": getCode,
	}
	for name, fn := range ajaxFuncs {
		router.Add("GET", "/"+name, fn)
	}
}

func getUsers(req *http.Request) ([]byte, error) {
	projectId, _, e := getProjectId(req)
	if e != nil {
		return nil, e
	}
	u, e := users(projectId)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

func getTools(req *http.Request) ([]byte, error) {
	projectId, _, e := getProjectId(req)
	if e != nil {
		return nil, e
	}
	t, e := tools(projectId)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tools": t})
}

func getCode(req *http.Request) ([]byte, error) {
	resultId, _, e := getId(req, "resultid", "result")
	if e != nil {
		return nil, e
	}
	r, e := db.ToolResult(bson.M{db.ID: resultId}, bson.M{db.FILEID: 1})
	if e != nil {
		return nil, e
	}
	f, e := db.File(bson.M{db.ID: r.GetFileId()}, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(f.Data)})
}

func getSkeletons(req *http.Request) ([]byte, error) {
	projectId, _, e := getProjectId(req)
	if e != nil {
		return nil, e
	}
	vals, e := skeletons(projectId)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"skeletons": vals})
}

func getChart(req *http.Request) ([]byte, error) {
	subId, _, e := getSubId(req)
	if e != nil {
		return nil, e
	}
	n, e := GetString(req, "file")
	if e != nil {
		return nil, e
	}
	r, e := GetString(req, "result")
	if e != nil {
		return nil, e
	}
	switch r {
	case jacoco.NAME:
		cId, e := util.ReadId(req.FormValue("childfileid"))
		if e != nil {
			return nil, e
		}
		r += "-" + cId.Hex()
	case junit.NAME:
		r, _ = util.Extension(n)
		if cId, e := util.ReadId(req.FormValue("childfileid")); e == nil {
			r += "-" + cId.Hex()
		}
	}
	files, e := db.Files(bson.M{db.SUBID: subId, db.NAME: n}, bson.M{db.DATA: 0})
	if e != nil {
		return nil, e
	}
	c, e := LoadChart(r, files)
	if e != nil {
		return nil, e
	}
	return util.JSON(c)
}
