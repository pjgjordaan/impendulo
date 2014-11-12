package ajax

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/calc"
	"github.com/godfried/impendulo/web/calc/tables"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

func Table(r *http.Request) ([]byte, error) {
	if e := r.ParseForm(); e != nil {
		return nil, e
	}
	t, e := webutil.String(r, "type")
	if e != nil {
		return nil, e
	}
	switch t {
	case "assignment":
		return assignmentTable(r)
	case "submission":
		return submissionTable(r)
	case "overview":
		return overviewTable(r)
	case "file":
		return fileTable(r)
	default:
		return nil, fmt.Errorf("unsupported type %s", t)
	}
}

func fileTable(r *http.Request) ([]byte, error) {
	sid, e := webutil.Id(r, "submission-id")
	if e != nil {
		return nil, e
	}
	tm, e := Metrics(r, calc.FILE)
	if e != nil {
		return nil, e
	}
	c, e := db.Table(sid.Hex(), calc.FILE)
	if e == nil {
		return util.JSON(map[string]interface{}{"table-data": c.Data, "table-fields": tables.FileFields(), "table-metrics": tm})
	}
	f, e := db.FileInfos(bson.M{db.SUBID: sid, db.TYPE: bson.M{db.IN: []project.Type{project.SRC, project.TEST}}})
	if e != nil {
		return nil, e
	}
	td, e := tables.File(f, tm)
	if e != nil {
		return nil, e
	}
	if e := db.AddTable(td, calc.FILE, sid.Hex()); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"table-data": td, "table-fields": tables.FileFields(), "table-metrics": tm})
}

func submissionTable(r *http.Request) ([]byte, error) {
	var id string
	m := bson.M{}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		m[db.PROJECTID] = pid
		id = pid.Hex()
	}
	if uid, e := webutil.String(r, "user-id"); e == nil {
		m[db.USER] = uid
		id = uid
	}
	if aid, e := webutil.Id(r, "assignment-id"); e == nil {
		m[db.ASSIGNMENTID] = aid
		id = aid.Hex()
	}
	tm, e := Metrics(r, calc.SUBMISSION)
	if e != nil {
		return nil, e
	}
	c, e := db.Table(id, calc.SUBMISSION)
	if e == nil {
		return util.JSON(map[string]interface{}{"table-data": c.Data, "table-fields": tables.SubmissionFields(), "table-metrics": tm})
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	td, e := tables.Submission(s, tm)
	if e != nil {
		return nil, e
	}
	if e := db.AddTable(td, calc.SUBMISSION, id); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"table-data": td, "table-fields": tables.SubmissionFields(), "table-metrics": tm})
}

func assignmentTable(r *http.Request) ([]byte, error) {
	t, e := webutil.String(r, "assignment-type")
	if e != nil {
		return nil, e
	}
	id, e := webutil.String(r, "id")
	if e != nil {
		return nil, e
	}
	tm, e := Metrics(r, calc.ASSIGNMENT)
	if e != nil {
		return nil, e
	}
	c, e := db.Table(id, calc.ASSIGNMENT)
	if e == nil {
		return util.JSON(map[string]interface{}{"table-data": c.Data, "table-fields": tables.AssignmentFields(), "table-metrics": tm})
	}
	m := bson.M{}
	switch t {
	case "project":
		pid, e := convert.Id(id)
		if e != nil {
			return nil, e
		}
		m[db.PROJECTID] = pid
	case "user":
		aids, e := db.UserAssignmentIds(id)
		if e != nil {
			return nil, e
		}
		m[db.ID] = bson.M{db.IN: aids}
	default:
		return nil, fmt.Errorf("invalid assignment type %s", t)
	}
	a, e := db.Assignments(m, nil)
	if e != nil {
		return nil, e
	}
	td, e := tables.Assignment(a, tm)
	if e != nil {
		return nil, e
	}
	if e := db.AddTable(td, calc.ASSIGNMENT, id); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"table-data": td, "table-fields": tables.AssignmentFields(), "table-metrics": tm})
}

func overviewTable(r *http.Request) ([]byte, error) {
	v, e := webutil.String(r, "view")
	if e != nil {
		return nil, e
	}
	tm, e := Metrics(r, calc.OVERVIEW)
	if e != nil {
		return nil, e
	}
	c, e := db.Table(v, calc.OVERVIEW)
	if e == nil {
		return util.JSON(map[string]interface{}{"table-data": c.Data, "table-fields": tables.OverviewFields(v), "table-metrics": tm})
	}
	var td tables.T
	switch v {
	case "user":
		u, e := db.Users(nil)
		if e != nil {
			return nil, e
		}
		if td, e = tables.User(u, tm); e != nil {
			return nil, e
		}
	case "project":
		p, e := db.Projects(nil, nil)
		if e != nil {
			return nil, e
		}
		if td, e = tables.Project(p, tm); e != nil {
			return nil, e
		}
	default:
		return nil, fmt.Errorf("unknown view %s", v)
	}
	if e := db.AddTable(td, calc.OVERVIEW, v); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"table-data": td, "table-fields": tables.OverviewFields(v), "table-metrics": tm})
}
