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

package charts

import (
	"errors"
	"fmt"
	"math"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/context"
	"labix.org/v2/mgo/bson"
)

type (
	//Chart represents the x and y values used to draw the charts.
	C struct {
		sid, user, id string
		result        *context.Result
		start         int64
		keys          map[string]string
		Data          D
	}

	D      []map[string]interface{}
	avgVal struct {
		count int
		val   *tool.ChartVal
	}
)

var (
	NoFilesError       = errors.New("no files to load chart for")
	NoSubmissionsError = errors.New("no submissions to create chart for")
	NoValuesError      = errors.New("no values found")
)

func Tool(r *context.Result, files []*project.File) (D, error) {
	if len(files) == 0 {
		return nil, NoFilesError
	}
	s, e := db.Submission(bson.M{db.ID: files[0].SubId}, nil)
	if e != nil {
		return nil, e
	}
	c := New(s, r)
	for _, f := range files {
		f, e = db.File(bson.M{db.ID: f.Id}, nil)
		if e != nil {
			continue
		}
		switch c.result.Type {
		case tool.SUMMARY:
			addAll(c, f)
		default:
			addSingle(c, f)
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

func addAll(c *C, f *project.File) {
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

func addSingle(c *C, f *project.File) {
	if _, e := convert.Id(f.Results[c.result.Raw()]); e != nil {
		return
	}
	r, e := db.ChartResult(bson.M{db.ID: f.Results[c.result.Raw()]}, nil)
	if e != nil {
		return
	}
	c.Add(f.Time, r.ChartVals())
	return
}

//Add inserts new coordinates into data used to display a chart.
func (c *C) Add(t int64, vs []*tool.ChartVal) {
	if len(vs) == 0 {
		return
	}
	x := util.Round(float64(t-c.start)/1000.0, 2)
	title := c.user + " \u2192 " + c.result.Format()
	r := c.result.Raw()
	for i, v := range vs {
		p := map[string]interface{}{
			"x": x, "y": v.Y, "key": c.Key(v.Name), "name": v.Name,
			"groupid": c.id, "user": c.user, "title": title,
			"created": c.start, "time": t, "pos": i, "sid": c.sid, "rid": r,
		}
		c.Data = append(c.Data, p)
	}
}

func (c *C) Key(n string) string {
	k, ok := c.keys[n]
	if ok {
		return k
	}
	c.keys[n] = bson.NewObjectId().Hex()
	return c.keys[n]
}

func (c *C) Len() int {
	return len(c.Data)
}
func (c *C) Swap(i, j int) {
	c.Data[i], c.Data[j] = c.Data[j], c.Data[i]
}
func (c *C) Less(i, j int) bool {
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

//New initialises new chart data.
func New(s *project.Submission, result *context.Result) *C {
	return &C{
		keys:   make(map[string]string),
		user:   s.User,
		id:     bson.NewObjectId().Hex(),
		result: result,
		start:  s.Time,
		Data:   NewData(),
		sid:    s.Id.Hex(),
	}
}

func NewData() D {
	return make(D, 0, 1000)
}

func User() (D, error) {
	us, e := db.Users(nil)
	if e != nil {
		return nil, e
	}
	d := NewData()
	for _, u := range us {
		c := db.TypeCounts(u.Name)
		p := map[string]interface{}{
			"key": u.Name, "submissions": c[0],
			"snapshots": c[1], "launches": c[2],
		}
		d = append(d, p)
	}
	return d, nil
}

func Project() (D, error) {
	ps, e := db.Projects(bson.M{}, nil)
	if e != nil {
		return nil, e
	}
	d := NewData()
	for _, p := range ps {
		c := db.TypeCounts(p.Id)
		v := map[string]interface{}{
			"key": p.Name, "submissions": c[0],
			"snapshots": c[1], "launches": c[2],
			"id": p.Id,
		}
		d = append(d, v)
	}
	return d, nil
}

type scoreFunc func(*project.Submission, *context.Result) (*tool.ChartVal, error)

func Submission(subs []*project.Submission, r *context.Result, score string) (D, error) {
	if len(subs) == 0 {
		return nil, NoSubmissionsError
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
	d := NewData()
	for _, s := range subs {
		n, e := db.ProjectName(s.ProjectId)
		if e != nil {
			continue
		}
		sc, e := db.FileCount(s.Id, project.SRC)
		if e != nil {
			continue
		}
		lc, e := db.FileCount(s.Id, project.LAUNCH)
		if e != nil {
			continue
		}
		v, e := f(s, r)
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
	prev := -1
	outliers := make(map[int]util.E, len(d))
	var stdDev, mean float64
	for len(outliers)-prev > 0 {
		mean = calcMean(d, outliers)
		stdDev = calcStdDeviation(d, outliers, mean)
		prev = len(outliers)
		addOutliers(d, outliers, mean, stdDev)
	}
	max := mean + 3*stdDev
	min := mean - 3*stdDev
	for i, _ := range outliers {
		y := d[i]["y"].(float64)
		if y < min {
			d[i]["y"] = int(min)
		} else if y > max {
			d[i]["y"] = int(max)
		}
		d[i]["outlier"] = y
	}
	return d, nil
}

func calcMean(vals []map[string]interface{}, outliers map[int]util.E) float64 {
	mean := 0.0
	for i, c := range vals {
		if _, ok := outliers[i]; ok {
			continue
		}
		mean += c["y"].(float64)
	}
	mean /= float64(len(vals) - len(outliers))
	return mean
}

func calcStdDeviation(vals []map[string]interface{}, outliers map[int]util.E, mean float64) float64 {
	stdDev := 0.0
	for i, c := range vals {
		if _, ok := outliers[i]; ok {
			continue
		}
		stdDev += math.Pow(c["y"].(float64)-mean, 2.0)
	}
	stdDev = math.Sqrt(stdDev / float64(len(vals)-len(outliers)))
	return stdDev
}

func addOutliers(vals []map[string]interface{}, outliers map[int]util.E, mean, stdDev float64) {
	max := mean + 3*stdDev
	min := mean - 3*stdDev
	for i, c := range vals {
		if _, ok := outliers[i]; ok {
			continue
		}
		y := c["y"].(float64)
		if y < min || y > max {
			outliers[i] = util.E{}
		}
	}
}

func norm(y, m, d float64) float64 {
	return (1.0 / (d * math.Sqrt(2*math.Pi))) * math.Exp(-math.Pow(y-m, 2.0)/(2*math.Pow(d, 2.0)))
}

func finalScore(s *project.Submission, r *context.Result) (*tool.ChartVal, error) {
	f, rid, e := lastInfo(s.Id, r)
	if e != nil {
		return nil, e
	}
	t := (f.Time - s.Time) / 1000.0
	return firstVal(rid, t)
}

func firstVal(rid bson.ObjectId, t int64) (*tool.ChartVal, error) {
	r, e := db.ChartResult(bson.M{db.ID: rid}, nil)
	if e != nil {
		return nil, e
	}
	vs := r.ChartVals()
	if len(vs) == 0 || vs[0] == nil {
		return nil, NoValuesError
	}
	vs[0].X = t
	return vs[0], nil
}

func (a *avgVal) add(v *tool.ChartVal) {
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

func (a *avgVal) chartVal() *tool.ChartVal {
	a.val.Y = util.Round(a.val.Y/float64(a.count), 2)
	return a.val
}

func averageScore(s *project.Submission, r *context.Result) (*tool.ChartVal, error) {
	fs, e := db.Files(bson.M{db.SUBID: s.Id, db.TYPE: project.SRC}, bson.M{db.DATA: 0}, 0)
	if e != nil {
		return nil, e
	}
	if len(fs) == 0 {
		return nil, fmt.Errorf("no src files in submission %s", s.Id.Hex())
	}
	var ts []*project.File
	if r.Name != "" && db.Contains(db.TESTS, bson.M{db.NAME: r.Name + ".java", db.TYPE: junit.USER}) {
		ts, e = db.Files(bson.M{db.SUBID: s.Id, db.NAME: r.Name + ".java", db.TYPE: project.TEST}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
		if e != nil {
			return nil, e
		}
	}
	cv := new(avgVal)
	for i, f := range fs {
		ft := (f.Time - s.Time) / 1000.0
		if id, e := convert.GetId(fs[i].Results, r.Raw()); e == nil {
			v, e := firstVal(id, ft)
			if e == nil {
				cv.add(v)
			}
			continue
		}
		for _, t := range ts {
			if id, e := convert.GetId(f.Results, r.Raw()+"-"+t.Id.Hex()); e == nil {
				v, e := firstVal(id, ft)
				if e != nil {
					continue
				}
				cv.add(v)
				r.FileID = t.Id
				break
			}
		}
	}
	if cv.val == nil {
		return nil, fmt.Errorf("no chartvals for %s", r.Format())
	}
	return cv.chartVal(), nil
}

func lastInfo(sid bson.ObjectId, r *context.Result) (*project.File, bson.ObjectId, error) {
	fs, e := db.Files(bson.M{db.SUBID: sid, db.TYPE: project.SRC}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
	if e != nil {
		return nil, "", e
	}
	if len(fs) == 0 {
		return nil, "", fmt.Errorf("no src files in submission %s", sid.Hex())
	}
	var ts []*project.File
	if r.Name != "" && db.Contains(db.TESTS, bson.M{db.NAME: r.Name + ".java", db.TYPE: junit.USER}) {
		ts, e = db.Files(bson.M{db.SUBID: sid, db.NAME: r.Name + ".java", db.TYPE: project.TEST}, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
		if e != nil {
			return nil, "", e
		}
	}
	for i, f := range fs {
		if id, e := convert.GetId(fs[i].Results, r.Raw()); e == nil {
			return fs[i], id, nil
		}
		for _, t := range ts {
			if id, e := convert.GetId(f.Results, r.Raw()+"-"+t.Id.Hex()); e == nil {
				return fs[i], id, nil
			}
		}
	}
	return nil, "", fmt.Errorf("no results found for %s", r.Format())
}
