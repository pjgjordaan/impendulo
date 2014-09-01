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
	"math"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/stats"
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
	I      map[string]string
	avgVal struct {
		count int
		val   *result.ChartVal
	}
)

var (
	NoFilesError       = errors.New("no files to load chart for")
	NoSubmissionsError = errors.New("no submissions to create chart for")
	NoAssignmentsError = errors.New("no assignments to create chart for")
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
		addSingle(c, f)
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
		v := []*result.ChartVal{&result.ChartVal{Name: "Launches", Y: 0.0, FileId: l.Id}}
		c.Add(l.Time, v)
	}
	return c.Data, nil
}

func addAll(c *C, f *project.File) {
	for _, id := range f.Results {
		if _, e := convert.Id(id); e != nil {
			continue
		}
		r, e := db.Charter(bson.M{db.ID: id}, nil)
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
	r, e := db.Charter(bson.M{db.ID: f.Results[c.result.Raw()]}, nil)
	if e != nil {
		return
	}
	c.Add(f.Time, r.ChartVals())
	return
}

//Add inserts new coordinates into data used to display a chart.
func (c *C) Add(t int64, vs []*result.ChartVal) {
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
		vals := stats.TypeCounts(u.Name)
		vals["key"] = u.Name
		d = append(d, vals)
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
		vals := stats.TypeCounts(p.Id)
		vals["key"] = p.Name
		vals["id"] = p.Id
		d = append(d, vals)
	}
	return d, nil
}

func Assignment(as []*project.Assignment, x *context.Result, y *context.Result) (D, I, error) {
	if len(as) == 0 {
		return nil, nil, NoAssignmentsError
	}
	d := NewData()
	calc := stats.NewCalc()
	names := make(map[bson.ObjectId]string)
	info := I{"x": x.Format(), "y": y.Format()}
	for _, a := range as {
		n, ok := names[a.ProjectId]
		if !ok {
			var e error
			if n, e = db.ProjectName(a.ProjectId); e != nil {
				continue
			}
			names[a.ProjectId] = n
		}
		subs, e := db.Submissions(bson.M{db.ASSIGNMENTID: a.Id}, nil)
		if e != nil {
			continue
		}
		count, xTotal, yTotal := 0, 0.0, 0.0
		for _, s := range subs {
			xVal, xN, e := calc.Calc(x, s)
			if e != nil {
				util.Log(e)
				continue
			}
			yVal, yN, e := calc.Calc(y, s)
			if e != nil {
				util.Log(e)
				continue
			}
			count++
			xTotal += xVal
			yTotal += yVal
			if _, ok := info["x-unit"]; !ok {
				info["x-unit"] = xN
			}
			if _, ok := info["y-unit"]; !ok {
				info["y-unit"] = yN
			}
		}
		if count == 0 {
			continue
		}
		p := map[string]interface{}{
			"key": a.Id.Hex(), "x": xTotal / float64(count), "y": yTotal / float64(count),
			"project": n, "name": a.Name,
		}
		d = append(d, p)
	}
	AddOutliers(d)
	return d, info, nil
}

func AddOutliers(d D) {
	prev := -1
	outliers := make(map[int]util.E, len(d))
	for len(outliers)-prev > 0 {
		mean := mean(d, outliers)
		stdDev := stdDeviation(d, outliers, mean)
		prev = len(outliers)
		addOutliers(d, outliers, mean, stdDev)
	}
	if len(outliers) == 0 {
		return
	}
	mean := mean(d, outliers)
	stdDev := stdDeviation(d, outliers, mean)
	n := float64(len(d))
	inv := 0.5 * (2.82843 * stdDev * util.ErfInverse((n-0.5)/n, 100))
	min := mean - inv
	max := mean + inv
	for i, _ := range outliers {
		y := d[i]["y"].(float64)
		if y < min {
			d[i]["y"] = int(min)
		} else if y > max {
			d[i]["y"] = int(max)
		}
		d[i]["outlier"] = y
	}
}

func Submission(subs []*project.Submission, x *context.Result, y *context.Result) (D, I, error) {
	if len(subs) == 0 {
		return nil, nil, NoSubmissionsError
	}
	d := NewData()
	calc := stats.NewCalc()
	names := make(map[bson.ObjectId]string)
	info := I{"x": x.Format(), "y": y.Format()}
	for _, s := range subs {
		n, ok := names[s.ProjectId]
		if !ok {
			var e error
			if n, e = db.ProjectName(s.ProjectId); e != nil {
				continue
			}
			names[s.ProjectId] = n
		}
		xVal, xN, e := calc.Calc(x, s)
		if e != nil {
			util.Log(e)
			continue
		}
		yVal, yN, e := calc.Calc(y, s)
		if e != nil {
			util.Log(e)
			continue
		}
		if _, ok := info["x-unit"]; !ok {
			info["x-unit"] = xN
		}
		if _, ok := info["y-unit"]; !ok {
			info["y-unit"] = yN
		}
		p := map[string]interface{}{
			"key": s.Id.Hex(), "user": s.User, "x": xVal, "y": yVal,
			"project": n,
		}
		d = append(d, p)
	}
	AddOutliers(d)
	return d, info, nil
}

func mean(vals []map[string]interface{}, outliers map[int]util.E) float64 {
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

func stdDeviation(vals []map[string]interface{}, outliers map[int]util.E, mean float64) float64 {
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
	for i, c := range vals {
		if _, ok := outliers[i]; ok {
			continue
		}
		y := c["y"].(float64)
		z := math.Abs(mean-y) / stdDev
		p := (1 - math.Erf(z/math.Sqrt(2))) * float64(len(vals))
		if p < 0.5 {
			outliers[i] = util.E{}
		}
	}
}
