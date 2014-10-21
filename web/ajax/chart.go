package ajax

import (
	"fmt"

	"labix.org/v2/mgo/bson"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/charts"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"

	"net/http"
	"sort"
	"strings"
)

func ChartOptions(r *http.Request) ([]byte, error) {
	var rs []string
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		rs = db.ProjectResults(pid)
	} else if u, e := webutil.String(r, "user-id"); e == nil {
		rs = db.UserResults(u)
	} else {
		rs = db.AllResults()
	}
	sort.Strings(rs)
	other := []string{"Time", "Lines", util.Title(project.SRC.String()), util.Title(project.LAUNCH.String()), util.Title(project.TEST.String()), "Testcases", "Passed"}
	ops := make(Selects, len(rs)+len(other))
	for i, o := range other {
		ops[i] = &Select{Id: o, Name: o}
	}
	for i, r := range rs {
		ops[i+len(other)] = &Select{Id: r, Name: strings.Replace(r, ":", " \u2192 ", -1)}
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
	x, e := webutil.String(r, "x")
	if e != nil {
		return nil, e
	}
	xd, e := context.NewResult(x)
	if e != nil {
		return nil, e
	}
	y, e := webutil.String(r, "y")
	if e != nil {
		return nil, e
	}
	yd, e := context.NewResult(y)
	v, e := webutil.String(r, "view")
	if e != nil {
		return nil, e
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
	return util.JSON(map[string]interface{}{"chart-data": d, "chart-info": i})
}

func assignmentChart(r *http.Request) ([]byte, error) {
	x, e := webutil.String(r, "x")
	if e != nil {
		return nil, e
	}
	xd, e := context.NewResult(x)
	if e != nil {
		return nil, e
	}
	y, e := webutil.String(r, "y")
	if e != nil {
		return nil, e
	}
	yd, e := context.NewResult(y)
	t, e := webutil.String(r, "assignment-type")
	if e != nil {
		return nil, e
	}
	id, e := webutil.String(r, "id")
	if e != nil {
		return nil, e
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
	return util.JSON(map[string]interface{}{"chart-data": d, "chart-info": i})
}

func submissionChart(r *http.Request) ([]byte, error) {
	x, e := webutil.String(r, "x")
	if e != nil {
		return nil, e
	}
	xd, e := context.NewResult(x)
	if e != nil {
		return nil, e
	}
	y, e := webutil.String(r, "y")
	if e != nil {
		return nil, e
	}
	yd, e := context.NewResult(y)
	if e != nil {
		return nil, e
	}
	t, e := webutil.String(r, "submission-type")
	if e != nil {
		return nil, e
	}
	id, e := webutil.String(r, "id")
	if e != nil {
		return nil, e
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
		m[db.USER] = id
	default:
		return nil, fmt.Errorf("invalid submission chart type %s", t)
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	d, i, e := charts.Submission(s, xd, yd)
	if e != nil {
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
	rn, e := webutil.String(r, "result")
	if e != nil {
		return nil, e
	}
	rd, e := context.NewResult(rn)
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

func _fileChart(s, fn string, r *context.Result) (charts.D, error) {
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
	r, e := context.NewResult(cmp)
	if e != nil {
		return nil, e
	}
	fs, e := db.Files(bson.M{db.NAME: fn, db.SUBID: sid}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	return charts.Tool(r, fs)
}
