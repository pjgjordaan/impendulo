package ajax

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/stats"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

func FileNames(r *http.Request) ([]byte, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return nil, e
	}
	ns, e := db.ProjectFileNames(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"filenames": ns})
}

func Assignments(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if id, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = id
	}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		m[db.PROJECTID] = pid
	}
	sm := bson.M{}
	if t, e := webutil.Int64(r, "max-start"); e == nil && t > 0 {
		sm[db.LT] = t
	}
	if t, e := webutil.Int64(r, "min-start"); e == nil && t > 0 {
		sm[db.GT] = t
	}
	if len(sm) > 0 {
		m[db.START] = sm
	}
	em := bson.M{}
	if t, e := webutil.Int64(r, "max-end"); e == nil && t > 0 {
		em[db.LT] = t
	}
	if t, e := webutil.Int64(r, "min-end"); e == nil && t > 0 {
		em[db.GT] = t
	}
	if len(em) > 0 {
		m[db.START] = em
	}
	as, e := db.Assignments(m, nil)
	if e != nil {
		return nil, e
	}
	if r.FormValue("counts") != "true" {
		return util.JSON(map[string]interface{}{"assignments": as})
	}
	cs := make(map[string]map[string]interface{})
	pc := stats.ProjectTestCases(as[0].ProjectId)
	for _, a := range as {
		fmt.Println(a)
		cs[a.Id.Hex()] = assignmentCounts(a.Id, pc)
	}
	return util.JSON(map[string]interface{}{"assignments": as, "counts": cs})
}

//Projects loads a list of projects.
func Projects(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = pid
	}
	p, e := db.Projects(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"projects": p})
}

func Files(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if sid, e := webutil.Id(r, "submission-id"); e == nil {
		m[db.SUBID] = sid
	}
	if id, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = id
	}
	format, _ := webutil.String(r, "format")
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

func Submissions(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if sid, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = sid
	}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		m[db.PROJECTID] = pid
	}
	if aid, e := webutil.Id(r, "assignment-id"); e == nil {
		m[db.ASSIGNMENTID] = aid
	}
	tm := bson.M{}
	if start, e := webutil.Int64(r, "time-start"); e == nil && start > 0 {
		tm[db.GT] = start
	}
	if end, e := webutil.Int64(r, "time-end"); e == nil && end > 0 {
		tm[db.LT] = end
	}
	if len(tm) > 0 {
		m[db.TIME] = tm
	}
	ss, e := db.Submissions(m, nil)
	if e != nil || len(ss) == 0 {
		return nil, e
	}
	if r.FormValue("counts") != "true" {
		return util.JSON(map[string]interface{}{"submissions": ss})
	}
	cs := make(map[string]map[string]interface{})
	pc := stats.ProjectTestCases(ss[0].ProjectId)
	for _, s := range ss {
		cs[s.Id.Hex()] = submissionCounts(s.Id, pc)
	}
	return util.JSON(map[string]interface{}{"submissions": ss, "counts": cs})
}

func Langs(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"langs": []tool.Language{tool.JAVA, tool.C}})
}

func Skeletons(r *http.Request) ([]byte, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return nil, e
	}
	s, e := db.Skeletons(bson.M{db.PROJECTID: pid}, bson.M{db.DATA: 0}, db.NAME)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"skeletons": s})
}
