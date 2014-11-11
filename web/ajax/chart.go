package ajax

import (
	"fmt"

	"labix.org/v2/mgo/bson"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/result/description"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/calc"
	"github.com/godfried/impendulo/web/calc/charts"
	"github.com/godfried/impendulo/web/webutil"

	"net/http"
)

func chartOptions(r *http.Request) ([]byte, error) {
	t, e := webutil.String(r, "type")
	if e != nil {
		return nil, e
	}
	l, e := calc.ParseLevel(t)
	if e != nil {
		return nil, e
	}
	ops, e := Metrics(r, l)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"options": ops})
}

func Chart(r *http.Request) ([]byte, error) {
	if e := r.ParseForm(); e != nil {
		return nil, e
	}
	t, e := webutil.String(r, "type")
	if e != nil {
		return nil, e
	}
	switch t {
	case "file":
		return fileChart(r)
	case "submission":
		return submissionChart(r)
	case "assignment":
		return assignmentChart(r)
	case "overview":
		return overviewChart(r)
	default:
		return nil, fmt.Errorf("unsupported chart type %s", t)
	}
}

func overviewChart(r *http.Request) ([]byte, error) {
	xd, e := webutil.Description(r, "x")
	if e != nil {
		return nil, e
	}
	yd, e := webutil.Description(r, "y")
	if e != nil {
		return nil, e
	}
	v, e := webutil.String(r, "view")
	if e != nil {
		return nil, e
	}
	c, e := db.Chart(v, calc.OVERVIEW, xd.Raw(), yd.Raw())
	if e == nil {
		return util.JSON(map[string]interface{}{"chart-data": c.Data, "chart-info": c.Info})
	}
	var d charts.D
	var i charts.I
	switch v {
	case "user":
		u, e := db.Users(nil)
		if e != nil {
			return nil, e
		}
		if d, i, e = charts.User(u, xd, yd); e != nil {
			return nil, e
		}
	case "project":
		p, e := db.Projects(nil, nil)
		if e != nil {
			return nil, e
		}
		if d, i, e = charts.Project(p, xd, yd); e != nil {
			return nil, e
		}
	default:
		return nil, fmt.Errorf("unknown view %s", v)
	}
	if e := db.AddChart(d, i, calc.OVERVIEW, v, xd.Raw(), yd.Raw()); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"chart-data": d, "chart-info": i})
}

func assignmentChart(r *http.Request) ([]byte, error) {
	xd, e := webutil.Description(r, "x")
	if e != nil {
		return nil, e
	}
	yd, e := webutil.Description(r, "y")
	if e != nil {
		return nil, e
	}
	t, e := webutil.String(r, "assignment-type")
	if e != nil {
		return nil, e
	}
	id, e := webutil.String(r, "id")
	if e != nil {
		return nil, e
	}
	c, e := db.Chart(id, calc.ASSIGNMENT, xd.Raw(), yd.Raw())
	if e == nil {
		return util.JSON(map[string]interface{}{"chart-data": c.Data, "chart-info": c.Info})
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
		return nil, fmt.Errorf("invalid submission chart type %s", t)
	}
	a, e := db.Assignments(m, nil)
	if e != nil {
		return nil, e
	}
	d, i, e := charts.Assignment(a, xd, yd)
	if e != nil {
		return nil, e
	}
	if e := db.AddChart(d, i, calc.ASSIGNMENT, id, xd.Raw(), yd.Raw()); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"chart-data": d, "chart-info": i})
}

func submissionChart(r *http.Request) ([]byte, error) {
	xd, e := webutil.Description(r, "x")
	if e != nil {
		return nil, e
	}
	yd, e := webutil.Description(r, "y")
	if e != nil {
		return nil, e
	}
	m := bson.M{}
	var id string
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
	c, e := db.Chart(id, calc.SUBMISSION, xd.Raw(), yd.Raw())
	if e == nil {
		return util.JSON(map[string]interface{}{"chart-data": c.Data, "chart-info": c.Info})
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	d, i, e := charts.Submission(s, xd, yd)
	if e != nil {
		return nil, e
	}
	if e := db.AddChart(d, i, calc.SUBMISSION, id, xd.Raw(), yd.Raw()); e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"chart-data": d, "chart-info": i})
}

func fileChart(r *http.Request) ([]byte, error) {
	sid, e := webutil.Id(r, "submission-id")
	if e != nil {
		return nil, e
	}
	fn, e := webutil.String(r, "file")
	if e != nil {
		return nil, e
	}
	rd, e := webutil.Description(r, "result")
	if e != nil {
		return nil, e
	}
	subs, e := webutil.Strings(r, "submissions[]")
	if e != nil {
		return nil, e
	}
	cmps, e := webutil.Strings(r, "comparables[]")
	if e != nil {
		return nil, e
	}
	var d charts.D
	for _, s := range subs {
		if c, e := _fileChart(s, fn, rd); e != nil {
			util.Log(e)
		} else {
			d = append(d, c...)
		}
	}
	for _, cmp := range cmps {
		if c, e := _cmpChart(sid, cmp, fn); e != nil {
			util.Log(e)
		} else {
			d = append(d, c...)
		}
	}
	return util.JSON(map[string]interface{}{"chart": d})
}

func _fileChart(s, fn string, r *description.D) (charts.D, error) {
	id, e := convert.Id(s)
	if e != nil {
		return nil, e
	} else if r.FileID != "" {
		r.FileID = db.UserTestId(id)
	}
	fs, e := db.Files(bson.M{db.NAME: fn, db.SUBID: id}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	return charts.Tool(r, fs)
}

func _cmpChart(sid bson.ObjectId, cmp, fn string) (charts.D, error) {
	r, e := description.New(cmp)
	if e != nil {
		return nil, e
	}
	fs, e := db.Files(bson.M{db.NAME: fn, db.SUBID: sid}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	return charts.Tool(r, fs)
}
