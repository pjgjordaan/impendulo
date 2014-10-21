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

	"github.com/godfried/impendulo/user"

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
	I      map[string]interface{}
	avgVal struct {
		count int
		val   *result.ChartVal
	}
	O map[int]util.E
)

var (
	NoFilesError       = errors.New("no files to load chart for")
	NoSubmissionsError = errors.New("no submissions to create chart for")
	NoAssignmentsError = errors.New("no assignments to create chart for")
	NoProjectsError    = errors.New("no projects to create chart for")
	NoUsersError       = errors.New("no users to create chart for")
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

func User(us []*user.User, x *context.Result, y *context.Result) (D, I, error) {
	if len(us) == 0 {
		return nil, nil, NoUsersError
	}
	d := NewData()
	c := stats.NewCalc()
	i := I{"x": x.Format(), "y": y.Format()}
	sumX, sumY := 0.0, 0.0
	for _, u := range us {
		if xTotal, yTotal, e := c.User(u, x, y); e == nil {
			sumX += xTotal
			sumY += yTotal
			v := map[string]interface{}{
				"key": "assignmentsview?user-id=" + u.Name, "x": xTotal, "y": yTotal,
				"title": u.Name,
			}
			d = append(d, v)
		}
	}
	i["x-unit"] = c.XN
	i["y-unit"] = c.YN
	d.AddOutliers("x")
	d.AddOutliers("y")
	i.AddStats(d, sumX, sumY)
	return d, i, nil
}

func Project(ps []*project.P, x *context.Result, y *context.Result) (D, I, error) {
	if len(ps) == 0 {
		return nil, nil, NoProjectsError
	}
	d := NewData()
	c := stats.NewCalc()
	i := I{"x": x.Format(), "y": y.Format()}
	sumX, sumY := 0.0, 0.0
	for _, p := range ps {
		if xTotal, yTotal, e := c.Project(p, x, y); e == nil {
			sumX += xTotal
			sumY += yTotal
			v := map[string]interface{}{
				"url": "assignmentschart?project-id=" + p.Id.Hex(), "x": xTotal, "y": yTotal,
				"title": p.Name,
			}
			d = append(d, v)
		}
	}
	i["x-unit"] = c.XN
	i["y-unit"] = c.YN
	d.AddOutliers("x")
	d.AddOutliers("y")
	i.AddStats(d, sumX, sumY)
	return d, i, nil
}

func Assignment(as []*project.Assignment, x *context.Result, y *context.Result) (D, I, error) {
	if len(as) == 0 {
		return nil, nil, NoAssignmentsError
	}
	d := NewData()
	c := stats.NewCalc()
	names := make(map[bson.ObjectId]string)
	i := I{"x": x.Format(), "y": y.Format()}
	sumX, sumY := 0.0, 0.0
	for _, a := range as {
		n, ok := names[a.ProjectId]
		if !ok {
			var e error
			if n, e = db.ProjectName(a.ProjectId); e != nil {
				continue
			}
			names[a.ProjectId] = n
		}
		if xTotal, yTotal, e := c.Assignment(a, x, y); e == nil {
			sumX += xTotal
			sumY += yTotal
			p := map[string]interface{}{
				"url": "submissionschart?assignment-id=" + a.Id.Hex(), "x": xTotal, "y": yTotal,
				"title": n + " " + a.Name,
			}
			d = append(d, p)
		}
	}
	i["x-unit"] = c.XN
	i["y-unit"] = c.YN
	d.AddOutliers("x")
	d.AddOutliers("y")
	i.AddStats(d, sumX, sumY)
	return d, i, nil
}

func Submission(subs []*project.Submission, x *context.Result, y *context.Result) (D, I, error) {
	if len(subs) == 0 {
		return nil, nil, NoSubmissionsError
	}
	d := NewData()
	c := stats.NewCalc()
	names := make(map[bson.ObjectId]string)
	i := I{"x": x.Format(), "y": y.Format()}
	sumX, sumY := 0.0, 0.0
	for _, s := range subs {
		n, ok := names[s.ProjectId]
		if !ok {
			var e error
			if n, e = db.ProjectName(s.ProjectId); e != nil {
				continue
			}
			names[s.ProjectId] = n
		}
		if xVal, yVal, e := c.Submission(s, x, y); e == nil {
			sumX += xVal
			sumY += yVal
			p := map[string]interface{}{
				"url": "filesview?submission-id=" + s.Id.Hex(), "title": s.User + "'s " + n, "x": xVal, "y": yVal,
			}
			d = append(d, p)
		}
	}
	i["x-unit"] = c.XN
	i["y-unit"] = c.YN
	d.AddOutliers("x")
	d.AddOutliers("y")
	i.AddStats(d, sumX, sumY)
	return d, i, nil
}

func (i I) AddStats(d D, sumX, sumY float64) {
	mx := util.Round(sumX/float64(len(d)), 2)
	my := util.Round(sumY/float64(len(d)), 2)
	sx, sy := d.StandardDeviations(mx, my)
	mxo, myo := d.AdjustedMeans()
	sxo, syo := d.AdjustedStandardDeviations(mxo, myo)
	i["mean"] = bson.M{"x": mx, "y": my, "x_o": mxo, "y_o": myo, "title": "Mean"}
	i["stddev"] = bson.M{"x": mx, "y": my, "x_o": mxo, "y_o": myo, "rx": sx, "ry": sy, "rx_o": sxo, "ry_o": syo}
}

func (d D) StandardDeviations(mx, my float64) (float64, float64) {
	sx, sy := 0.0, 0.0
	for _, v := range d {
		if f, e := convert.GetFloat64(v, "x"); e == nil {
			sx += math.Pow(f-mx, 2.0)
		}
		if f, e := convert.GetFloat64(v, "y"); e == nil {
			sy += math.Pow(f-my, 2.0)
		}
	}
	sx = util.Round(math.Sqrt(sx/float64(len(d))), 2.0)
	sy = util.Round(math.Sqrt(sy/float64(len(d))), 2.0)
	return sx, sy
}

func (d D) AdjustedMeans() (float64, float64) {
	mx, my := 0.0, 0.0
	for _, v := range d {
		if f, e := convert.GetFloat64(v, "x_o"); e == nil {
			mx += f
		} else if f, e := convert.GetFloat64(v, "x"); e == nil {
			mx += f
		}
		if f, e := convert.GetFloat64(v, "y_o"); e == nil {
			my += f
		} else if f, e := convert.GetFloat64(v, "y"); e == nil {
			my += f
		}
	}
	mx = util.Round(mx/float64(len(d)), 2.0)
	my = util.Round(my/float64(len(d)), 2.0)
	return mx, my
}

func (d D) AdjustedStandardDeviations(mx, my float64) (float64, float64) {
	sx, sy := 0.0, 0.0
	for _, v := range d {
		if f, e := convert.GetFloat64(v, "x_o"); e == nil {
			sx += math.Pow(f-mx, 2.0)
		} else if f, e := convert.GetFloat64(v, "x"); e == nil {
			sx += math.Pow(f-mx, 2.0)
		}
		if f, e := convert.GetFloat64(v, "y_o"); e == nil {
			sy += math.Pow(f-my, 2.0)
		} else if f, e := convert.GetFloat64(v, "y"); e == nil {
			sy += math.Pow(f-my, 2.0)
		}
	}
	sx = util.Round(math.Sqrt(sx/float64(len(d))), 2.0)
	sy = util.Round(math.Sqrt(sy/float64(len(d))), 2.0)
	return sx, sy
}

func (d D) AddOutliers(axis string) {
	prev := -1
	o := make(O, len(d))
	for len(o)-prev > 0 {
		mean := o.Mean(d, axis)
		stdDev := o.StandardDeviation(d, mean, axis)
		prev = len(o)
		o.Add(d, mean, stdDev, axis)
	}
	if len(o) == 0 {
		return
	}
	mean := o.Mean(d, axis)
	stdDev := o.StandardDeviation(d, mean, axis)
	n := float64(len(d))
	inv := 0.5 * (2.82843 * stdDev * util.ErfInverse((n-0.5)/n, 100))
	min := mean - inv
	max := mean + inv
	for i, _ := range o {
		v := d[i][axis].(float64)
		if v < min {
			d[i][axis+"_o"] = int(min)
		} else if v > max {
			d[i][axis+"_o"] = int(max)
		}
	}
}

func (o O) Mean(d D, axis string) float64 {
	mean := 0.0
	for i, c := range d {
		if _, ok := o[i]; ok {
			continue
		}
		mean += c[axis].(float64)
	}
	mean /= float64(len(d) - len(o))
	return mean
}

func (o O) StandardDeviation(d D, mean float64, axis string) float64 {
	stdDev := 0.0
	for i, c := range d {
		if _, ok := o[i]; ok {
			continue
		}
		stdDev += math.Pow(c[axis].(float64)-mean, 2.0)
	}
	stdDev = math.Sqrt(stdDev / float64(len(d)-len(o)))
	return stdDev
}

func (o O) Add(d D, mean, stdDev float64, axis string) {
	for i, c := range d {
		if _, ok := o[i]; ok {
			continue
		}
		v := c[axis].(float64)
		z := math.Abs(mean-v) / stdDev
		p := (1 - math.Erf(z/math.Sqrt(2))) * float64(len(d))
		if p < 0.5 {
			o[i] = util.E{}
		}
	}
}
