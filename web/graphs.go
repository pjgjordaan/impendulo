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
	"github.com/godfried/impendulo/tool/diff"
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
	sub, err := db.Submission(bson.M{db.ID: files[0].SubId}, nil)
	if err != nil {
		return nil, err
	}
	chart := NewChart(sub)
	prev, err := db.File(bson.M{db.ID: files[0].Id}, nil)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		matcher := bson.M{db.ID: f.Id}
		cur, err := db.File(matcher, nil)
		if err != nil {
			continue
		}
		switch resultName {
		case tool.SUMMARY:
			err = addAll(chart, cur, prev)
		default:
			err = addSingle(chart, cur, prev, resultName)
		}
		if err != nil {
			return nil, err
		}
		prev = cur
	}
	file, err := db.File(bson.M{db.ID: files[0].Id}, bson.M{db.SUBID: 1})
	if err != nil {
		util.Log(err)
		return chart.Data, nil
	}
	matcher := bson.M{db.SUBID: file.SubId, db.TYPE: project.LAUNCH}
	launches, err := db.Files(matcher, nil, db.TIME)
	if err != nil {
		util.Log(err)
		return chart.Data, nil
	}
	for i, launch := range launches {
		prev := launches[util.Max(i-1, 0)]
		val := []*tool.ChartVal{&tool.ChartVal{"Launches", 0.0, false, launch.Id}}
		err = chart.Add(launch, prev, val)
		if err != nil {
			return nil, err
		}
	}
	return chart.Data, nil
}

func addAll(chart *Chart, cur, prev *project.File) error {
	for name, id := range cur.Results {
		result, err := db.ChartResult(name, bson.M{db.ID: id}, nil)
		if err != nil {
			continue
		}
		err = chart.Add(cur, prev, result.ChartVals()[:1])
		if err != nil {
			return err
		}
	}
	return nil
}

func addSingle(chart *Chart, cur, prev *project.File, resultName string) error {
	result, err := db.ChartResult(resultName, bson.M{db.ID: cur.Results[resultName]}, nil)
	if err != nil {
		return nil
	}
	return chart.Add(cur, prev, result.ChartVals())
}

func SubmissionChart(subs []*project.Submission) (ret ChartData) {
	ret = NewChartData()
	for _, sub := range subs {
		snapshots, err := fileCount(sub.Id, project.SRC)
		if err != nil {
			continue
		}
		launches, err := fileCount(sub.Id, project.LAUNCH)
		if err != nil {
			continue
		}
		name, err := projectName(sub.ProjectId)
		if err != nil {
			continue
		}
		matcher := bson.M{db.TYPE: project.SRC, db.SUBID: sub.Id}
		selector := bson.M{db.TIME: 1}
		lastFile, err := db.LastFile(matcher, selector)
		if err != nil {
			continue
		}
		time := (lastFile.Time - sub.Time) / 1000.0
		point := map[string]interface{}{
			"snapshots": snapshots, "launches": launches, "project": name,
			"key": sub.Id.Hex(), "user": sub.User, "status": sub.Status,
			"description": sub.Result(), "time": time,
		}
		ret = append(ret, point)
	}
	return
}

func overviewChart(tipe string) (ret ChartData, err error) {
	switch tipe {
	case "user":
		ret, err = UserChart()
	case "project":
		ret, err = ProjectChart()
	default:
		return nil, fmt.Errorf("Unknown type %s.", tipe)
	}
	return

}

func UserChart() (ret ChartData, err error) {
	ret = NewChartData()
	users, err := db.Users(bson.M{})
	if err != nil {
		return
	}
	for _, u := range users {
		counts := TypeCounts(u.Name)
		point := map[string]interface{}{
			"key": u.Name, "submissions": counts[0],
			"snapshots": counts[1], "launches": counts[2],
		}
		ret = append(ret, point)
	}
	return
}

func ProjectChart() (ret ChartData, err error) {
	ret = NewChartData()
	projects, err := db.Projects(bson.M{}, nil)
	if err != nil {
		return
	}
	for _, p := range projects {
		counts := TypeCounts(p.Id)
		point := map[string]interface{}{
			"key": p.Name, "submissions": counts[0],
			"snapshots": counts[1], "launches": counts[2],
			"id": p.Id,
		}
		ret = append(ret, point)
	}
	return
}

func TypeCounts(id interface{}) (counts []int) {
	counts = []int{0, 0, 0}
	var matcher string
	switch id.(type) {
	case string:
		matcher = db.USER
	case bson.ObjectId:
		matcher = db.PROJECTID
	default:
		return
	}
	subs, err := db.Submissions(bson.M{matcher: id}, nil)
	if err != nil {
		return
	}
	counts[0] = len(subs)
	if counts[0] == 0 {
		return
	}
	for _, sub := range subs {
		if s, serr := fileCount(sub.Id, project.SRC); serr == nil {
			counts[1] += s
		}
		if l, lerr := fileCount(sub.Id, project.LAUNCH); lerr == nil {
			counts[2] += l
		}
	}
	return
}

//Add inserts new coordinates into data used to display a chart.
func (this *Chart) Add(curFile, prevFile *project.File, vals []*tool.ChartVal) (err error) {
	if len(vals) == 0 {
		return
	}
	x := float64(curFile.Time-this.start) / 1000
	var d string
	hasDiff := curFile.Id != prevFile.Id && curFile.Type == project.SRC
	if hasDiff {
		cur := diff.NewResult(curFile)
		prev := diff.NewResult(prevFile)
		d, err = prev.Create(cur)
		if err != nil {
			return
		}
	}
	for _, curVal := range vals {
		point := map[string]interface{}{
			"x": x, "y": curVal.Y, "key": curVal.Name + " " + this.subId,
			"name": curVal.Name, "subId": this.subId, "user": this.user,
			"created": this.start, "time": curFile.Time, "show": curVal.Show,
		}
		if hasDiff {
			point["diff"] = d
		}
		this.Data = append(this.Data, point)
	}
	return
}

//NewChart initialises new chart data.
func NewChart(submission *project.Submission) *Chart {
	return &Chart{
		start: submission.Time,
		user:  submission.User,
		subId: submission.Id.Hex(),
		Data:  NewChartData(),
	}
}

func NewChartData() ChartData {
	return make(ChartData, 0, 1000)
}
