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
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Chart represents the x and y values used to draw the charts.
	Chart struct {
		start int64
		user  string
		subId string
		Data  ChartData
	}

	ChartData []map[string]interface{}
)

func LoadChart(resultName string, files []*project.File) (ChartData, error) {
	if len(files) == 0 {
		return nil, errors.New("No files to load chart for.")
	}
	s, e := db.Submission(bson.M{db.ID: files[0].SubId}, nil)
	if e != nil {
		return nil, e
	}
	c := NewChart(s)
	for _, f := range files {
		f, e = db.File(bson.M{db.ID: f.Id}, nil)
		if e != nil {
			continue
		}
		switch resultName {
		case tool.SUMMARY:
			e = addAll(c, f)
		default:
			e = addSingle(c, f, resultName)
		}
		if e != nil {
			return nil, e
		}
	}
	f, e := db.File(bson.M{db.ID: files[0].Id}, bson.M{db.SUBID: 1})
	if e != nil {
		util.Log(e)
		return c.Data, nil
	}
	ls, e := db.Files(bson.M{db.SUBID: f.SubId, db.TYPE: project.LAUNCH}, nil, db.TIME)
	if e != nil {
		util.Log(e)
		return c.Data, nil
	}
	for _, l := range ls {
		v := []*tool.ChartVal{&tool.ChartVal{"Launches", 0.0, false, l.Id}}
		c.Add(l, v)
	}
	return c.Data, nil
}

func addAll(c *Chart, f *project.File) error {
	for _, id := range f.Results {
		r, e := db.ChartResult(bson.M{db.ID: id}, nil)
		if e != nil {
			continue
		}
		c.Add(f, r.ChartVals()[:1])
	}
	return nil
}

func addSingle(c *Chart, f *project.File, n string) error {
	r, e := db.ChartResult(bson.M{db.ID: f.Results[n]}, nil)
	if e != nil {
		return nil
	}
	c.Add(f, r.ChartVals())
	return nil
}

func SubmissionChart(subs []*project.Submission) ChartData {
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
		n, e := projectName(s.ProjectId)
		if e != nil {
			continue
		}
		f, e := db.LastFile(bson.M{db.TYPE: project.SRC, db.SUBID: s.Id}, bson.M{db.TIME: 1})
		if e != nil {
			continue
		}
		t := (f.Time - s.Time) / 1000.0
		p := map[string]interface{}{
			"snapshots": sc, "launches": lc, "project": n,
			"key": s.Id.Hex(), "user": s.User, "status": s.Status,
			"description": s.Result(), "time": t,
		}
		d = append(d, p)
	}
	return d
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
func (c *Chart) Add(f *project.File, vs []*tool.ChartVal) {
	if len(vs) == 0 {
		return
	}
	x := util.Round(float64(f.Time-c.start)/1000.0, 2)
	for _, v := range vs {
		p := map[string]interface{}{
			"x": x, "y": v.Y, "key": v.Name + " " + c.subId,
			"name": v.Name, "subId": c.subId, "user": c.user,
			"created": c.start, "time": f.Time, "show": v.Show,
		}
		c.Data = append(c.Data, p)
	}
}

//NewChart initialises new chart data.
func NewChart(s *project.Submission) *Chart {
	return &Chart{
		start: s.Time,
		user:  s.User,
		subId: s.Id.Hex(),
		Data:  NewChartData(),
	}
}

func NewChartData() ChartData {
	return make(ChartData, 0, 1000)
}
