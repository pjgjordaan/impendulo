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
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

//LoadChart calculates GraphArgs for a given result.
func LoadChart(result string, files []*project.File) (chart tool.Chart) {
	switch result {
	case tool.CODE:
	case tool.SUMMARY:
	default:
		chart = loadChart(result, files)
	}
	return
}

func loadChart(name string, files []*project.File) (chart tool.Chart) {
	if len(files) == 0 {
		return
	}
	chart = tool.NewChart()
	selector := bson.M{project.TIME: 1, project.RESULTS: 1}
	for _, f := range files {
		matcher := bson.M{project.ID: f.Id}
		file, err := db.File(matcher, selector)
		if err != nil {
			continue
		}
		result, err := db.ChartResult(name,
			bson.M{project.ID: file.Results[name]}, nil)
		if err != nil {
			util.Log(err, file.Results[name])
			continue
		}
		chart.Add(float64(file.Time), result.ChartVals())
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
		chart.Add(float64(launch.Time), map[string]float64{"Launch": 0.0})
	}
	return
}
