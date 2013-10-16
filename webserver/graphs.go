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
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

func LoadChart(resultName string, files []*project.File, startTime int64) (chart tool.Chart) {
	if len(files) == 0 {
		return
	}
	chart = tool.NewChart()
	selector := bson.M{project.TIME: 1, project.RESULTS: 1}
	adjust := float64(startTime - files[0].Time)
	sub, err := db.Submission(bson.M{project.ID: files[0].SubId}, nil)
	if err != nil {
		util.Log(err)
		return
	}
	for _, f := range files {
		matcher := bson.M{project.ID: f.Id}
		file, err := db.File(matcher, selector)
		if err != nil {
			continue
		}
		result, err := db.ChartResult(resultName,
			bson.M{project.ID: file.Results[resultName]}, nil)
		if err != nil {
			continue
		}
		chart.Add(sub, float64(file.Time), adjust, result.ChartVals())
	}
	file, err := db.File(bson.M{project.ID: files[0].Id}, bson.M{project.SUBID: 1})
	if err != nil {
		util.Log(err)
		return
	}
	matcher := bson.M{project.SUBID: file.SubId, project.TYPE: project.LAUNCH}
	launches, err := db.Files(matcher, nil, project.TIME)
	if err != nil {
		util.Log(err)
		return
	}
	for _, launch := range launches {
		chart.Add(sub, float64(launch.Time), adjust, []tool.ChartVal{{"Launches", 0.0, false}})
	}
	return
}

func SubmissionChart(subs []*project.Submission) (ret tool.Chart) {
	ret = tool.NewChart()
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
		ret.Data = append(ret.Data, point)
	}
	return
}

func overviewChart(tipe string) (ret []map[string]interface{}, err error) {
	var chart tool.Chart
	switch tipe {
	case "user":
		chart, err = UserChart()
	case "project":
		chart, err = ProjectChart()
	default:
		return nil, fmt.Errorf("Unknown type %s.", tipe)
	}
	if err == nil {
		ret = chart.Data
	}
	return

}

func UserChart() (ret tool.Chart, err error) {
	ret = tool.NewChart()
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
		ret.Data = append(ret.Data, point)
	}
	return
}

func ProjectChart() (ret tool.Chart, err error) {
	ret = tool.NewChart()
	projects, err := db.Projects(bson.M{}, nil)
	if err != nil {
		return
	}
	for _, p := range projects {
		counts := TypeCounts(p.Id)
		point := map[string]interface{}{
			"key": p.Name, "submissions": counts[0],
			"snapshots": counts[1], "launches": counts[2],
		}
		ret.Data = append(ret.Data, point)
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
