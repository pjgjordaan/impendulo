package ajax

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/stats"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

func FileInfos(r *http.Request) ([]byte, error) {
	sid, e := webutil.Id(r, "submission-id")
	if e != nil {
		return nil, e
	}
	fs, e := db.FileInfos(bson.M{db.SUBID: sid, db.TYPE: bson.M{db.IN: []project.Type{project.SRC, project.TEST}}})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"fileinfos": fs})
}

func FileNames(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		ss, e := db.IDs(db.SUBMISSIONS, bson.M{db.PROJECTID: pid})
		if e != nil {
			return nil, e
		}
		m[db.SUBID] = bson.M{db.IN: ss}
	}
	if sid, e := webutil.Id(r, "submission-id"); e == nil {
		m[db.SUBID] = sid
	}
	ns, e := db.FileNames(m)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"filenames": ns})
}

func BasicFileInfos(r *http.Request) ([]byte, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return nil, e
	}
	fs, e := db.ProjectBasicFileInfos(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"fileinfos": fs})
}

func Assignments(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if id, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = id
	}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		m[db.PROJECTID] = pid
	}
	if u, e := webutil.String(r, "user-id"); e == nil {
		aids, e := db.UserAssignmentIds(u)
		if e != nil {
			return nil, e
		}
		m[db.ID] = bson.M{db.IN: aids}
	}
	if sm := loadTimes(r, "start"); len(sm) > 0 {
		m[db.START] = sm
	}
	if em := loadTimes(r, "end"); len(em) > 0 {
		m[db.END] = em
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
		cs[a.Id.Hex()] = assignmentCounts(a.Id, pc)
	}
	return util.JSON(map[string]interface{}{"assignments": as, "counts": cs})
}

func loadTimes(r *http.Request, n string) bson.M {
	m := bson.M{}
	if t, e := webutil.Int64(r, "max-"+n); e == nil && t > 0 {
		m[db.LT] = t
	}
	if t, e := webutil.Int64(r, "min-"+n); e == nil && t > 0 {
		m[db.GT] = t
	}
	return m
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
	if u, e := webutil.String(r, "user-id"); e == nil {
		m[db.USER] = u
	}
	if aid, e := webutil.Id(r, "assignment-id"); e == nil {
		m[db.ASSIGNMENTID] = aid
	}
	if tm := loadTimes(r, "time"); len(tm) > 0 {
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
	m := bson.M{}
	if id, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = id
	}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		m[db.PROJECTID] = pid
	}
	s, e := db.Skeletons(m, bson.M{db.DATA: 0}, db.NAME)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"skeletons": s})
}
