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

package webserver

import (
	"errors"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

func LoadChart(resultName string, files []*project.File, startTime int64) (tool.ChartData, error) {
	if len(files) == 0 {
		return nil, errors.New("No files to load chart for.")
	}
	selector := bson.M{project.TIME: 1, project.RESULTS: 1}
	sub, err := db.Submission(bson.M{project.ID: files[0].SubId}, nil)
	if err != nil {
		return nil, err
	}
	chart := tool.NewChart(sub, float64(startTime-files[0].Time))
	for _, f := range files {
		matcher := bson.M{project.ID: f.Id}
		file, err := db.File(matcher, selector)
		if err != nil {
			continue
		}
		switch resultName {
		case tool.SUMMARY:
			addAll(chart, file)
		default:
			addSingle(chart, file, resultName)
		}
	}
	file, err := db.File(bson.M{project.ID: files[0].Id}, bson.M{project.SUBID: 1})
	if err != nil {
		util.Log(err)
		return chart.Data, nil
	}
	matcher := bson.M{project.SUBID: file.SubId, project.TYPE: project.LAUNCH}
	launches, err := db.Files(matcher, nil, project.TIME)
	if err != nil {
		util.Log(err)
		return chart.Data, nil
	}
	for _, launch := range launches {
		val := []tool.ChartVal{{"Launches", 0.0, false}}
		chart.Add(float64(launch.Time), val)
	}
	return chart.Data, nil
}

func addAll(chart *tool.Chart, file *project.File) {
	ftime := float64(file.Time)
	for name, id := range file.Results {
		result, err := db.ChartResult(name, bson.M{project.ID: id}, nil)
		if err != nil {
			continue
		}
		chart.Add(ftime, result.ChartVals(true))
	}
}

func addSingle(chart *tool.Chart, file *project.File, resultName string) {
	result, err := db.ChartResult(resultName,
		bson.M{project.ID: file.Results[resultName]}, nil)
	if err != nil {
		return
	}
	chart.Add(float64(file.Time), result.ChartVals(false))
}

func SubmissionChart(subs []*project.Submission) (ret tool.ChartData) {
	ret = tool.NewChartData()
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
		lastFile, err := db.LastFile(sub)
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

func overviewChart(tipe string) (ret tool.ChartData, err error) {
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

func UserChart() (ret tool.ChartData, err error) {
	ret = tool.NewChartData()
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

func ProjectChart() (ret tool.ChartData, err error) {
	ret = tool.NewChartData()
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
		matcher = project.USER
	case bson.ObjectId:
		matcher = project.PROJECT_ID
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
