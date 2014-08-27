//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package stats

import (
	"errors"
	"fmt"
	"strings"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/wc"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/context"

	"io/ioutil"

	"labix.org/v2/mgo/bson"

	"path/filepath"
)

type (
	Test struct {
		total, tests, errors, failures int
	}
	C struct {
		projects map[bson.ObjectId]int
	}
	Type string
)

const (
	TIME      Type = "time"
	LINES     Type = "lines"
	PASSED    Type = "passed"
	TESTCASES Type = "testcases"
)

var (
	NoValuesError = errors.New("no values found")
	NoTestsError  = errors.New("no tests for project")
)

func NewCalc() *C {
	return &C{projects: make(map[bson.ObjectId]int)}
}

func (c *C) Calc(rd *context.Result, s *project.Submission) (float64, string, error) {
	n := rd.Raw()
	if t, e := project.ParseType(n); e == nil {
		fc, e := db.FileCount(s.Id, t)
		if e != nil {
			return -1, "", e
		}
		return float64(fc), "files", nil
	} else if t, e := ParseType(n); e == nil {
		pt, ok := c.projects[s.ProjectId]
		if t == PASSED && !ok {
			pt = ProjectTestCases(s.ProjectId)
			c.projects[s.ProjectId] = pt
		}
		return calc(t, s, pt)
	} else {
		return Score(s, rd)
	}
}

func ParseType(n string) (Type, error) {
	n = strings.ToLower(n)
	switch n {
	case "time", "passed", "testcases", "lines":
		return Type(n), nil
	default:
		return Type(""), fmt.Errorf("unknown stats type %s", n)
	}
}

func calc(t Type, s *project.Submission, projectTests int) (float64, string, error) {
	switch t {
	case PASSED:
		st, e := TestStats(s.Id, projectTests)
		if e != nil {
			return -1, "", e
		}
		return st.Passed(), "%", nil
	case TESTCASES:
		c, e := SubmissionTestCases(s.Id)
		if e != nil {
			return -1, "", e
		}
		return float64(c), "", nil
	case TIME:
		d, e := Time(s)
		if e != nil {
			return -1, "", e
		}
		return float64(d), "seconds", nil
	case LINES:
		l, e := Lines(s.Id)
		if e != nil {
			return -1, "", e
		}
		return float64(l), "lines", nil
	default:
		return -1, "", fmt.Errorf("unknown calc type %s", t)
	}
}

func NewTest(total int) *Test {
	return &Test{total: total, tests: 0, errors: 0, failures: 0}
}

func (t *Test) Add(r *junit.Report) {
	t.tests += r.Tests
	t.errors += r.Errors
	t.failures += r.Failures
}

func (t *Test) Passed() float64 {
	if t.tests == 0 || t.tests > t.total || (t.errors+t.failures) > t.tests {
		return 0.0
	}
	return 100 * float64(t.total-t.errors-t.failures) / float64(t.total)
}

func TestStats(sid bson.ObjectId, projectTests int) (*Test, error) {
	if projectTests == 0 {
		return nil, NoTestsError
	}
	s, e := db.Submission(bson.M{db.ID: sid}, bson.M{db.PROJECTID: 1})
	if e != nil {
		return nil, e
	}
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: s.ProjectId, db.TYPE: bson.M{db.NE: junit.USER}}, bson.M{db.NAME: 1})
	if e != nil {
		return nil, e
	}
	if len(ts) == 0 {
		return nil, NoTestsError
	}
	n, _ := util.Extension(ts[0].Name)
	f, e := resultFile(sid, junit.NAME+":"+n)
	if e != nil {
		return nil, e
	}
	st := NewTest(projectTests)
	for _, t := range ts {
		n, _ = util.Extension(t.Name)
		id, e := convert.Id(f.Results[junit.NAME+":"+n])
		if e != nil {
			continue
		}
		r, e := db.JUnitResult(bson.M{db.ID: id}, nil)
		if e != nil {
			return nil, e
		}
		st.Add(r.Report)
	}
	return st, nil
}

func resultFile(sid bson.ObjectId, result string) (*project.File, error) {
	f, e := db.LastFile(bson.M{db.SUBID: sid, db.TYPE: project.SRC}, bson.M{db.RESULTS: 1})
	if e != nil {
		return nil, e
	}
	if _, e := convert.Id(f.Results[result]); e == nil {
		return f, nil
	}
	fs, e := db.Files(bson.M{db.SUBID: sid, db.TYPE: project.SRC}, bson.M{db.RESULTS: 1}, 0, "-"+db.TIME)
	if e != nil {
		return nil, e
	}
	for _, f := range fs {
		if _, e := convert.Id(f.Results[result]); e == nil {
			return f, nil
		}
	}
	return nil, fmt.Errorf("no file found with result %s for submission %s", result, sid)
}

func ProjectTestCases(pid bson.ObjectId) int {
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: pid, db.TYPE: bson.M{db.NE: junit.USER}}, nil)
	if e != nil {
		return 0
	}
	c := 0
	for _, t := range ts {
		if tc, e := TestCases(t); e != nil {
			util.Log(e)
		} else {
			c += tc
		}
	}
	return c
}

func TestCases(t *junit.Test) (int, error) {
	if t.TestCases > 0 {
		return t.TestCases, nil
	}
	p, e := config.JUNIT_TESTING.Path()
	if e != nil {
		return 0, e
	}
	td, e := ioutil.TempDir("", "")
	if e != nil {
		return 0, e
	}
	if e = util.Copy(td, p); e != nil {
		return 0, e
	}
	testTarget := tool.NewTarget(t.Name, t.Package, filepath.Join(td, t.Id.Hex()), tool.JAVA)
	if e = util.SaveFile(testTarget.FilePath(), t.Test); e != nil {
		return 0, e
	}
	if len(t.Data) != 0 {
		if e = util.Unzip(testTarget.PackagePath(), t.Data); e != nil {
			return 0, e
		}
	}
	sd, e := ioutil.TempDir("", "")
	if e != nil {
		return 0, e
	}
	sk, e := db.Skeleton(bson.M{db.PROJECTID: t.ProjectId}, nil)
	if e != nil {
		return 0, e
	}
	if e = util.Unzip(sd, sk.Data); e != nil {
		return 0, e
	}
	if t.Target.Dir, e = util.LocateDirectory(sd, "src"); e != nil {
		return -1, e
	}
	ju, e := junit.New(testTarget, t.Target, td, t.Id)
	if e != nil {
		return 0, e
	}
	r, e := ju.Run(t.Id, t.Target)
	if e != nil {
		return 0, e
	}
	c := r.(*junit.Result).Report.Tests
	return c, db.Update(db.TESTS, bson.M{db.ID: t.Id}, bson.M{db.SET: bson.M{db.TESTCASES: c}})
}

func SubmissionTestCases(sid bson.ObjectId) (int, error) {
	f, e := db.LastFile(bson.M{db.SUBID: sid, db.TYPE: project.TEST}, bson.M{db.TESTCASES: 1})
	if e != nil {
		return 0, nil
	}
	if f.TestCases > 0 {
		return f.TestCases, nil
	}
	c, e := FileTestCases(f.Id)
	if e != nil {
		return -1, e
	}
	return c, db.Update(db.FILES, bson.M{db.ID: f.Id}, bson.M{db.SET: bson.M{db.TESTCASES: c}})
}

func FileTestCases(id bson.ObjectId) (int, error) {
	f, e := db.File(bson.M{db.ID: id}, nil)
	if e != nil {
		return 0, e
	}
	p, e := config.JUNIT_TESTING.Path()
	if e != nil {
		return -1, e
	}
	td, e := ioutil.TempDir("", "")
	if e != nil {
		return -1, e
	}
	if e = util.Copy(td, p); e != nil {
		return -1, e
	}
	testTarget := tool.NewTarget(f.Name, f.Package, filepath.Join(td, f.Id.Hex()), tool.JAVA)
	if e = util.SaveFile(testTarget.FilePath(), f.Data); e != nil {
		return -1, e
	}
	sd, e := ioutil.TempDir("", "")
	if e != nil {
		return -1, e
	}
	s, e := db.Submission(bson.M{db.ID: f.SubId}, bson.M{db.PROJECTID: 1})
	if e != nil {
		return 0, e
	}
	sk, e := db.Skeleton(bson.M{db.PROJECTID: s.ProjectId}, nil)
	if e != nil {
		return 0, e
	}
	if e = util.Unzip(sd, sk.Data); e != nil {
		return -1, e
	}
	t, e := db.JUnitTest(bson.M{db.NAME: f.Name, db.PROJECTID: s.ProjectId}, bson.M{db.TARGET: 1})
	if e != nil {
		return -1, e
	}
	if t.Target.Dir, e = util.LocateDirectory(sd, "src"); e != nil {
		return -1, e
	}
	ju, e := junit.New(testTarget, t.Target, td, f.Id)
	if e != nil {
		return -1, e
	}
	r, e := ju.Run(f.Id, t.Target)
	if e != nil {
		return -1, e
	}
	return r.(*junit.Result).Report.Tests, nil
}

func Time(s *project.Submission) (int64, error) {
	f, e := db.LastFile(bson.M{db.SUBID: s.Id}, bson.M{db.TIME: 1})
	if e != nil {
		return -1, e
	}
	return (f.Time - s.Time) / 1000.0, nil
}

func Score(s *project.Submission, r *context.Result) (float64, string, error) {
	rid, e := lastResultId(s.Id, r)
	if e != nil {
		return -1, "", e
	}
	return score(rid)
}

func score(rid bson.ObjectId) (float64, string, error) {
	r, e := db.Charter(bson.M{db.ID: rid}, nil)
	if e != nil {
		return -1, "", e
	}
	vs := r.ChartVals()
	if len(vs) == 0 || vs[0] == nil {
		return -1, "", NoValuesError
	}
	return vs[0].Y, strings.ToLower(vs[0].Name), nil
}

func lastResultId(sid bson.ObjectId, r *context.Result) (bson.ObjectId, error) {
	fs, e := db.Files(bson.M{db.SUBID: sid, db.TYPE: project.SRC}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
	if e != nil {
		return "", e
	}
	if len(fs) == 0 {
		return "", fmt.Errorf("no src files in submission %s", sid.Hex())
	}
	var ts []*project.File
	if r.Name != "" && db.Contains(db.TESTS, bson.M{db.NAME: r.Name + ".java", db.TYPE: junit.USER}) {
		ts, e = db.Files(bson.M{db.SUBID: sid, db.NAME: r.Name + ".java", db.TYPE: project.TEST}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
		if e != nil {
			return "", e
		}
	}
	for i, f := range fs {
		if id, e := convert.GetId(fs[i].Results, r.Raw()); e == nil {
			return id, nil
		}
		for _, t := range ts {
			if id, e := convert.GetId(f.Results, r.Raw()+"-"+t.Id.Hex()); e == nil {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("no results found for %s", r.Format())
}

func Lines(sid bson.ObjectId) (int64, error) {
	f, e := db.LastFile(bson.M{db.SUBID: sid, db.TYPE: project.SRC}, bson.M{db.DATA: 1})
	if e != nil {
		return -1, nil
	}
	return wc.Lines(f.Data)
}

func TypeNames() []string {
	return []string{"assignments", "submissions", string(PASSED), string(project.SRC), string(project.LAUNCH), string(project.TEST)}
}

func TypeCounts(id interface{}) map[string]interface{} {
	c := make(map[string]interface{})
	testCases := 0
	var m string
	switch t := id.(type) {
	case string:
		m = db.USER
	case bson.ObjectId:
		m = db.PROJECTID
		testCases = ProjectTestCases(t)
	default:
		return c
	}
	ss, e := db.Submissions(bson.M{m: id}, nil)
	if e != nil {
		return c
	}
	c["submissions"] = len(ss)
	sc, lc, tc, ps := 0, 0, 0, 0.0
	pc := 0
	am := util.NewSet()
	for _, s := range ss {
		am.Add(s.AssignmentId.Hex())
		if c, e := db.FileCount(s.Id, project.SRC); e == nil {
			sc += c
		}
		if c, e := db.FileCount(s.Id, project.LAUNCH); e == nil {
			lc += c
		}
		if c, e := db.FileCount(s.Id, project.TEST); e == nil {
			tc += c
		}
		if m == db.USER {
			testCases = ProjectTestCases(s.ProjectId)
		}
		if c, _, e := calc(PASSED, s, testCases); e == nil {
			ps += c
			pc++
		}
	}
	c["assignments"] = len(am)
	if pc == 0 {
		c[string(PASSED)] = "N/A"
	} else {
		c[string(PASSED)] = util.Round(ps/float64(pc), 2)
	}
	c[string(project.SRC)] = sc
	c[string(project.LAUNCH)] = lc
	c[string(project.TEST)] = tc
	return c
}
