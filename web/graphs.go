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

package web

import (
	"errors"
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"labix.org/v2/mgo/bson"
)

type (
	//Chart represents the x and y values used to draw the charts.
	Chart struct {
		user, id, result string
		start            int64
		keys             map[string]string
		Data             ChartData
	}

	ChartData   []map[string]interface{}
	avgChartVal struct {
		count int
		val   *tool.ChartVal
	}
)

func LoadChart(rd *ResultDesc, files []*project.File) (ChartData, error) {
	if len(files) == 0 {
		return nil, errors.New("no files to load chart for")
	}
	s, e := db.Submission(bson.M{db.ID: files[0].SubId}, nil)
	if e != nil {
		return nil, e
	}
	c := NewChart(s, rd)
	for _, f := range files {
		f, e = db.File(bson.M{db.ID: f.Id}, nil)
		if e != nil {
			continue
		}
		switch rd.Type {
		case tool.SUMMARY:
			addAll(c, f)
		default:
			addSingle(c, f, rd)
		}
	}
	f, e := db.File(bson.M{db.ID: files[0].Id}, bson.M{db.SUBID: 1})
	if e != nil {
		util.Log(e)
		return c.Data, nil
	}
	ls, e := db.Files(bson.M{db.SUBID: f.SubId, db.TYPE: project.LAUNCH}, nil, 0, db.TIME)
	if e != nil {
		util.Log(e)
		return c.Data, nil
	}
	for _, l := range ls {
		v := []*tool.ChartVal{&tool.ChartVal{Name: "Launches", Y: 0.0, FileId: l.Id}}
		c.Add(l.Time, v)
	}
	return c.Data, nil
}

func addAll(c *Chart, f *project.File) {
	for _, id := range f.Results {
		if _, e := convert.Id(id); e != nil {
			continue
		}
		r, e := db.ChartResult(bson.M{db.ID: id}, nil)
		if e != nil {
			continue
		}
		c.Add(f.Time, r.ChartVals()[:1])
	}
	return
}

func addSingle(c *Chart, f *project.File, rd *ResultDesc) {
	if _, e := convert.Id(f.Results[rd.Raw()]); e != nil {
		return
	}
	r, e := db.ChartResult(bson.M{db.ID: f.Results[rd.Raw()]}, nil)
	if e != nil {
		return
	}
	c.Add(f.Time, r.ChartVals())
	return
}

type scoreFunc func(*project.Submission, *ResultDesc) (*tool.ChartVal, error)

func SubmissionChart(subs []*project.Submission, result *ResultDesc, score string) (ChartData, error) {
	if len(subs) == 0 {
		return nil, errors.New("no submissions to create chart for")
	}
	n, e := projectName(subs[0].ProjectId)
	if e != nil {
		return nil, e
	}
	var f scoreFunc
	switch score {
	case "final":
		f = finalScore
	case "average":
		f = averageScore
	default:
		return nil, fmt.Errorf("unsupported score type %s", score)
	}
	d := NewChartData()
	for _, s := range subs {
		sc, e := fileCount(s.Id, project.SRC)
		if e != nil {
			continue
		}
		lc, e := fileCount(s.Id, project.LAUNCH)
		if e != nil {
			continue
		}
		v, e := f(s, result)
		if e != nil {
			continue
		}
		p := map[string]interface{}{
			"snapshots": sc, "launches": lc, "project": n,
			"key": s.Id.Hex(), "user": s.User, "time": v.X, "y": v.Y,
			"description": v.Name,
		}
		d = append(d, p)
	}
	return d, nil
}

func finalScore(s *project.Submission, result *ResultDesc) (*tool.ChartVal, error) {
	f, rid, e := lastInfo(s.Id, result)
	if e != nil {
		return nil, e
	}
	t := (f.Time - s.Time) / 1000.0
	return firstChartVal(rid, t)
}

func firstChartVal(rid bson.ObjectId, t int64) (*tool.ChartVal, error) {
	r, e := db.ChartResult(bson.M{db.ID: rid}, nil)
	if e != nil {
		return nil, e
	}
	vs := r.ChartVals()
	if len(vs) == 0 || vs[0] == nil {
		return nil, errors.New("no values found")
	}
	vs[0].X = t
	return vs[0], nil
}

func (a *avgChartVal) add(v *tool.ChartVal) {
	if a.val == nil {
		a.count = 0
		a.val = v
	} else {
		a.val.Y += v.Y
		if v.X > a.val.X {
			a.val.X = v.X
		}
	}
	a.count++
}

func (a *avgChartVal) chartVal() *tool.ChartVal {
	a.val.Y /= float64(a.count)
	return a.val
}

func averageScore(s *project.Submission, rd *ResultDesc) (*tool.ChartVal, error) {
	fs, e := db.Files(bson.M{db.SUBID: s.Id, db.TYPE: project.SRC}, bson.M{db.DATA: 0}, 0)
	if e != nil {
		return nil, e
	}
	if len(fs) == 0 {
		return nil, fmt.Errorf("no src files in submission %s", s.Id.Hex())
	}
	var ts []*project.File
	if rd.Name != "" && db.Contains(db.TESTS, bson.M{db.NAME: rd.Name + ".java", db.TYPE: junit.USER}) {
		ts, e = db.Files(bson.M{db.SUBID: s.Id, db.NAME: rd.Name + ".java", db.TYPE: project.TEST}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
		if e != nil {
			return nil, e
		}
	}
	cv := new(avgChartVal)
	for i, f := range fs {
		ft := (f.Time - s.Time) / 1000.0
		if id, e := convert.GetId(fs[i].Results, rd.Raw()); e == nil {
			v, e := firstChartVal(id, ft)
			if e == nil {
				cv.add(v)
			}
			continue
		}
		for _, t := range ts {
			if id, e := convert.GetId(f.Results, rd.Raw()+"-"+t.Id.Hex()); e == nil {
				v, e := firstChartVal(id, ft)
				if e != nil {
					continue
				}
				cv.add(v)
				rd.FileID = t.Id
				break
			}
		}
	}
	return cv.chartVal(), nil
}

func lastInfo(sid bson.ObjectId, rd *ResultDesc) (*project.File, bson.ObjectId, error) {
	fs, e := db.Files(bson.M{db.SUBID: sid, db.TYPE: project.SRC}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
	if e != nil {
		return nil, "", e
	}
	if len(fs) == 0 {
		return nil, "", fmt.Errorf("no src files in submission %s", sid.Hex())
	}
	var ts []*project.File
	if rd.Name != "" && db.Contains(db.TESTS, bson.M{db.NAME: rd.Name + ".java", db.TYPE: junit.USER}) {
		ts, e = db.Files(bson.M{db.SUBID: sid, db.NAME: rd.Name + ".java", db.TYPE: project.TEST}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
		if e != nil {
			return nil, "", e
		}
	}
	for i, f := range fs {
		if id, e := convert.GetId(fs[i].Results, rd.Raw()); e == nil {
			return fs[i], id, nil
		}
		for _, t := range ts {
			if id, e := convert.GetId(f.Results, rd.Raw()+"-"+t.Id.Hex()); e == nil {
				return fs[i], id, nil
			}
		}
	}
	return nil, "", fmt.Errorf("no results found for %s", rd.Format())
}

func overviewChart(c string) (ChartData, error) {
	switch c {
	case "user":
		return UserChart()
	case "project":
		return ProjectChart()
	}
	return nil, fmt.Errorf("Unknown type %s.", c)
}

func UserChart() (ChartData, error) {
	us, e := db.Users(nil)
	if e != nil {
		return nil, e
	}
	d := NewChartData()
	for _, u := range us {
		c := TypeCounts(u.Name)
		p := map[string]interface{}{
			"key": u.Name, "submissions": c[0],
			"snapshots": c[1], "launches": c[2],
		}
		d = append(d, p)
	}
	return d, nil
}

func ProjectChart() (ChartData, error) {
	ps, e := db.Projects(bson.M{}, nil)
	if e != nil {
		return nil, e
	}
	d := NewChartData()
	for _, p := range ps {
		c := TypeCounts(p.Id)
		v := map[string]interface{}{
			"key": p.Name, "submissions": c[0],
			"snapshots": c[1], "launches": c[2],
			"id": p.Id,
		}
		d = append(d, v)
	}
	return d, nil
}

func TypeCounts(id interface{}) []int {
	c := []int{0, 0, 0}
	var m string
	switch id.(type) {
	case string:
		m = db.USER
	case bson.ObjectId:
		m = db.PROJECTID
	default:
		return c
	}
	ss, err := db.Submissions(bson.M{m: id}, nil)
	if err != nil {
		return c
	}
	c[0] = len(ss)
	if c[0] == 0 {
		return c
	}
	for _, s := range ss {
		if sc, e := fileCount(s.Id, project.SRC); e == nil {
			c[1] += sc
		}
		if l, e := fileCount(s.Id, project.LAUNCH); e == nil {
			c[2] += l
		}
	}
	return c
}

//Add inserts new coordinates into data used to display a chart.
func (c *Chart) Add(t int64, vs []*tool.ChartVal) {
	if len(vs) == 0 {
		return
	}
	x := util.Round(float64(t-c.start)/1000.0, 2)
	for i, v := range vs {
		p := map[string]interface{}{
			"x": x, "y": v.Y, "key": c.Key(v.Name), "name": v.Name,
			"groupid": c.id, "user": c.user, "title": c.user + " \u2192 " + c.result,
			"created": c.start, "time": t, "pos": i,
		}
		c.Data = append(c.Data, p)
	}
}

func (c *Chart) Key(n string) string {
	k, ok := c.keys[n]
	if ok {
		return k
	}
	c.keys[n] = bson.NewObjectId().Hex()
	return c.keys[n]
}

func (c *Chart) Len() int {
	return len(c.Data)
}
func (c *Chart) Swap(i, j int) {
	c.Data[i], c.Data[j] = c.Data[j], c.Data[i]
}
func (c *Chart) Less(i, j int) bool {
	if c.Data[i]["time"] == c.Data[j]["time"] {
		pi, ok := c.Data[i]["pos"].(int)
		if !ok {
			return false
		}
		pj, ok := c.Data[j]["pos"].(int)
		if !ok {
			return false
		}
		return pi <= pj
	}
	ti, ok := c.Data[i]["time"].(int64)
	if !ok {
		return false
	}
	tj, ok := c.Data[j]["time"].(int64)
	if !ok {
		return true
	}
	return ti <= tj
}

//NewChart initialises new chart data.
func NewChart(s *project.Submission, rd *ResultDesc) *Chart {
	return &Chart{
		keys:   make(map[string]string),
		user:   s.User,
		id:     bson.NewObjectId().Hex(),
		result: rd.Format(),
		start:  s.Time,
		Data:   NewChartData(),
	}
}

func NewChartData() ChartData {
	return make(ChartData, 0, 1000)
}

func hasChart(cs ...interface{}) bool {
	for _, c := range cs {
		if _, ok := c.(tool.ChartResult); ok {
			return true
		}
	}
	return false
}
