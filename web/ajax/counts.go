package ajax

import (
	"fmt"

	"labix.org/v2/mgo/bson"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/stats"
	"github.com/godfried/impendulo/web/webutil"

	"net/http"
)

func TypeCounts(r *http.Request) ([]byte, error) {
	v, e := webutil.String(r, "view")
	if e != nil {
		return nil, e
	}
	var d []map[string]interface{}
	switch v {
	case "user":
		us, e := db.Users(nil)
		if e != nil {
			return nil, e
		}
		d = make([]map[string]interface{}, len(us))
		for i, u := range us {
			vals := stats.TypeCounts(u.Name)
			vals["name"] = u.Name
			d[i] = vals
		}
	case "project":
		ps, e := db.Projects(nil, nil)
		if e != nil {
			return nil, e
		}
		d = make([]map[string]interface{}, len(ps))
		for i, p := range ps {
			vals := stats.TypeCounts(p.Id)
			vals["name"] = p.Name
			vals["id"] = p.Id.Hex()
			vals["lang"] = p.Lang
			vals["description"] = p.Description
			vals["time"] = p.Time
			d[i] = vals
		}
	default:
		return nil, fmt.Errorf("unknown view %s", v)
	}
	return util.JSON(map[string]interface{}{"typecounts": d, "categories": stats.TypeNames()})
}

func Counts(r *http.Request) ([]byte, error) {
	if sid, e := webutil.Id(r, "submission-id"); e == nil {
		s, e := db.Submission(bson.M{db.ID: sid}, bson.M{db.PROJECTID: 1})
		if e != nil {
			return nil, e
		}
		return util.JSON(map[string]interface{}{"counts": submissionCounts(sid, stats.ProjectTestCases(s.ProjectId))})
	}
	return nil, CountsError
}

func submissionCounts(sid bson.ObjectId, projectTests int) map[string]interface{} {
	c, e := stats.SubmissionTestCases(sid)
	if e != nil {
		util.Log(e)
	}
	st, e := stats.TestStats(sid, projectTests)
	if e != nil {
		util.Log(e)
		st = stats.NewTest(0)
	}
	counts := map[string]interface{}{"testcases": c, "passed": util.Round(st.Passed(), 2)}
	tipes := []project.Type{project.ALL, project.SRC, project.LAUNCH, project.TEST}
	for _, t := range tipes {
		m := bson.M{db.SUBID: sid}
		if t != project.ALL {
			m[db.TYPE] = t
		}
		c, e := db.Count(db.FILES, m)
		if e != nil {
			util.Log(e)
		}
		counts[t.String()] = c
	}
	return counts
}

func assignmentCounts(aid bson.ObjectId, projectTests int) map[string]interface{} {
	ss, e := db.Submissions(bson.M{db.ASSIGNMENTID: aid}, bson.M{db.ID: 1})
	if e != nil {
		util.Log(e)
		return nil
	}
	counts := map[string]interface{}{"submissions": len(ss)}
	pc := 0
	ps := string(stats.PASSED)
	for _, s := range ss {
		c := submissionCounts(s.Id, projectTests)
		for k, v := range c {
			switch t := v.(type) {
			case float64:
				if k == ps {
					pc++
				}
				if _, ok := counts[k]; !ok {
					counts[k] = t
				} else {
					counts[k] = t + counts[k].(float64)
				}
			case int:
				if _, ok := counts[k]; !ok {
					counts[k] = v
				} else {
					counts[k] = t + counts[k].(int)
				}
			default:
				util.Log(fmt.Errorf("cannot handle type %v", v))
			}
		}
	}
	if pc > 0 {
		counts[ps] = util.Round(counts[ps].(float64)/float64(pc), 2)
	} else {
		counts[ps] = "N/A"
	}
	return counts
}
