package ajax

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/tool/result/all"
	"github.com/godfried/impendulo/tool/result/description"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/calc"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
	"sort"
)

func ResultNames(r *http.Request) ([]byte, error) {
	sid, e := webutil.Id(r, "submission-id")
	if e != nil {
		return nil, e
	}
	f, e := webutil.String(r, "filename")
	if e != nil {
		return nil, e
	}
	rs, e := db.ResultNames(sid, f)
	if e != nil {
		util.Log(e)
		rs = db.BasicResultNames()
	}
	return util.JSON(map[string]interface{}{"resultnames": rs})
}

//Results retrieves the names  of all results found within a particular
//project or by a particular user.
func Results(r *http.Request) ([]byte, error) {
	var rs []string
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		rs = db.ProjectResults(pid)
	} else if u, e := webutil.String(r, "user-id"); e == nil {
		rs = db.UserResults(u)
	} else {
		return nil, ResultsError
	}
	s := make(description.Ds, len(rs))
	for i, r := range rs {
		var e error
		if s[i], e = description.New(r); e != nil {
			return nil, e
		}
	}
	sort.Sort(s)
	return util.JSON(map[string]interface{}{"results": s})
}

//Comparables retrieves other results which a given result
//can be compared to, i.e. different unit tests.
func Comparables(r *http.Request) ([]byte, error) {
	d, e := webutil.Description(r, "result")
	if e != nil {
		return nil, e
	}
	if d.Type != jacoco.NAME && d.Type != junit.NAME {
		return util.JSON(map[string]interface{}{"comparables": description.Ds{}})
	}
	f, e := webutil.File(r, "file-id", bson.M{db.DATA: 0})
	if e != nil {
		return nil, e
	}
	rid, e := convert.GetId(f.Results, d.Raw())
	if e != nil {
		return nil, e
	}
	tr, e := db.Tooler(bson.M{db.ID: rid}, nil)
	if e != nil {
		return nil, e
	}
	s, e := db.Submission(bson.M{db.ID: f.SubId}, bson.M{db.PROJECTID: 1})
	if e != nil {
		return nil, e
	}
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: s.ProjectId, db.NAME: bson.M{db.NE: tr.GetName() + ".java"}, db.TYPE: bson.M{db.NE: junit.USER}}, bson.M{db.NAME: 1, db.TYPE: 1})
	if e != nil {
		return nil, e
	}
	m := bson.M{db.SUBID: f.SubId, db.TYPE: project.TEST}
	if ut, e := db.File(bson.M{db.ID: tr.GetTestId()}, bson.M{db.NAME: 1}); e == nil {
		m[db.ID] = bson.M{db.NE: ut.Id}
	} else if !db.Contains(db.TESTS, bson.M{db.ID: tr.GetTestId()}) {
		return nil, e
	}
	uts, e := db.Files(m, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
	if e != nil {
		return nil, e
	}
	cmp := make(description.Ds, len(ts)+len(uts))
	for i, t := range ts {
		n, _ := util.Extension(t.Name)
		cmp[i] = &description.D{Type: tr.GetType(), Name: n}
	}
	for i, ut := range uts {
		n, _ := util.Extension(ut.Name)
		cmp[i+len(ts)] = &description.D{Type: tr.GetType(), Name: n, FileID: ut.Id}
	}
	return util.JSON(map[string]interface{}{"comparables": cmp})
}

func Comment(w http.ResponseWriter, r *http.Request) error {
	c, e := context.Load(r)
	if e != nil {
		return e
	}
	u, e := c.Username()
	if e != nil {
		return e
	}
	f, e := webutil.File(r, "file-id", bson.M{db.COMMENTS: 1})
	if e != nil {
		return e
	}
	start, e := convert.Int(r.FormValue("start"))
	if e != nil {
		return e
	}
	end, e := convert.Int(r.FormValue("end"))
	if e != nil {
		return e
	}
	d, e := webutil.String(r, "data")
	if e != nil {
		return e
	}
	f.Comments = append(f.Comments, &project.Comment{Data: d, User: u, Start: start, End: end})
	return db.Update(db.FILES, bson.M{db.ID: f.Id}, bson.M{db.SET: bson.M{db.COMMENTS: f.Comments}})
}

func Comments(r *http.Request) ([]byte, error) {
	c, e := commentor(r)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"comments": c.LoadComments()})
}

func commentor(r *http.Request) (project.Commentor, error) {
	if id, e := webutil.Id(r, "file-id"); e == nil {
		return db.File(bson.M{db.ID: id}, bson.M{db.COMMENTS: 1})
	} else if id, e := webutil.Id(r, "submission-id"); e == nil {
		return db.Submission(bson.M{db.ID: id}, bson.M{db.COMMENTS: 1})
	}
	return nil, CommentsError
}

func FileResults(r *http.Request) ([]byte, error) {
	f, e := webutil.File(r, "id", bson.M{db.RESULTS: 1})
	if e != nil {
		return nil, e
	}
	lines := make(map[string][]*result.Line, len(f.Results)*10)
	for _, v := range f.Results {
		rid, e := convert.Id(v)
		if e != nil {
			continue
		}
		r, e := db.Coder(bson.M{db.ID: rid}, nil)
		if e != nil {
			continue
		}
		lines[r.GetName()] = r.Lines()
	}
	return util.JSON(map[string]interface{}{"fileresults": lines})
}

//Code loads code for a given src file or test.
func Code(r *http.Request) ([]byte, error) {
	if tid, e := webutil.Id(r, "test-id"); e == nil {
		t, e := db.JUnitTest(bson.M{db.ID: tid}, bson.M{db.TEST: 1})
		if e != nil {
			return nil, e
		}
		return util.JSON(map[string]interface{}{"code": string(t.Test)})
	}
	m := bson.M{}
	if rid, e := webutil.Id(r, "result-id"); e == nil {
		tr, e := db.Tooler(bson.M{db.ID: rid}, bson.M{db.FILEID: 1})
		if e != nil {
			return nil, e
		}
		m[db.ID] = tr.GetFileId()
	}
	if id, e := webutil.Id(r, "file-id"); e == nil {
		m[db.ID] = id
	}
	if n, e := webutil.String(r, "tool-name"); e == nil {
		d, e := description.New(n)
		if e != nil {
			return nil, e
		}
		if d.FileID != "" {
			m[db.ID] = d.FileID
		} else if pid, e := webutil.Id(r, "project-id"); e == nil {
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

//loadTestCode loads a unit test's code
func loadTestCode(pid bson.ObjectId, n string) ([]byte, error) {
	t, e := db.JUnitTest(bson.M{db.PROJECTID: pid, db.NAME: n + ".java"}, bson.M{db.TEST: 1})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(t.Test)})
}

func Tests(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		m[db.PROJECTID] = pid
	}
	if id, e := webutil.Id(r, "id"); e == nil {
		m[db.ID] = id
	}
	t, e := db.JUnitTests(m, bson.M{db.TEST: 0, db.DATA: 0})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tests": t})
}

func TestTypes(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"types": junit.TestTypes()})
}

func TestData(r *http.Request) ([]byte, error) {
	c, e := context.Load(r)
	if e != nil {
		return nil, e
	}
	test, e := webutil.Description(r, "test")
	if e != nil {
		return nil, e
	}
	name, e := webutil.String(r, "data-name")
	if e != nil {
		return nil, e
	}
	tf, e := db.JUnitTest(bson.M{db.NAME: test.Name + ".java", db.PROJECTID: c.Browse.Pid}, bson.M{db.DATA: 1})
	if e != nil {
		return nil, e
	}
	d, e := ioutil.TempDir("", "")
	if e != nil {
		return nil, e
	}
	if e := util.Unzip(d, tf.Data); e != nil {
		return nil, e
	}
	defer os.RemoveAll(d)
	p, e := util.LocateFile(d, name)
	if e != nil {
		return nil, e
	}
	data, e := ioutil.ReadFile(p)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"data": string(data)})
}

func Metrics(r *http.Request, l calc.Level) (description.Ds, error) {
	var rs []string
	if pid, e := webutil.Id(r, "project-id"); e == nil {
		rs = db.ProjectResults(pid)
	} else if u, e := webutil.String(r, "user-id"); e == nil {
		rs = db.UserResults(u)
	} else {
		rs = db.AllResults()
	}
	var ops description.Ds
	switch l {
	case calc.OVERVIEW:
		ops = overviewMetrics()
	case calc.ASSIGNMENT:
		ops = assignmentMetrics()
	case calc.SUBMISSION:
		ops = submissionMetrics()
	default:
		return nil, fmt.Errorf("unsupported level %s", l)
	}
	for _, r := range rs {
		tipes, e := all.Types(r)
		if e != nil {
			return nil, e
		}
		for _, t := range tipes {
			id := r + "~" + t
			d, e := description.New(id)
			if e != nil {
				return nil, e
			}
			ops = append(ops, d)
		}
	}
	sort.Sort(ops)
	return ops, nil
}

func overviewMetrics() description.Ds {
	return append(assignmentMetrics(), description.Ds{{Type: "Submissions", Metric: "Average"}, {Type: "Assignments", Metric: "Total"}}...)
}

func assignmentMetrics() description.Ds {
	return append(submissionMetrics(), description.Ds{{Type: "Time", Metric: "Average"}, {Type: util.Title(project.SRC.String()), Metric: "Average"}, {Type: util.Title(project.LAUNCH.String()), Metric: "Average"}, {Type: util.Title(project.TEST.String()), Metric: "Average"}, {Type: "Testcases", Metric: "Average"}, {Type: "Submissions", Metric: "Total"}}...)
}

func submissionMetrics() description.Ds {
	return description.Ds{{Type: "Time", Metric: "Total"}, {Type: util.Title(project.SRC.String()), Metric: "Total"}, {Type: util.Title(project.LAUNCH.String()), Metric: "Total"}, {Type: util.Title(project.TEST.String()), Metric: "Total"}, {Type: "Testcases", Metric: "Total"}, {Type: "Passed", Metric: "Average"}}
}
