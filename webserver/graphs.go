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
	"labix.org/v2/mgo/bson"
)

//LoadChart calculates GraphArgs for a given result.
func LoadChart(result, tipe string, files []*project.File) (chart tool.Chart) {
	time := tipe == "time"
	switch result {
	/*case "All":
	graphData, max = loadAllGraphData(time, files)*/
	case tool.CODE:
	case tool.SUMMARY:
	default:
		chart = loadChart(result, files, time)
	}
	return
}

func loadChart(name string, files []*project.File, time bool) (chart tool.Chart) {
	for _, f := range files {
		file, err := db.File(bson.M{project.ID: f.Id}, bson.M{project.TIME: 1, project.RESULTS: 1})
		if err != nil {
			continue
		}
		result, err := db.ChartResult(name,
			bson.M{project.ID: file.Results[name]}, nil)
		if err != nil {
			continue
		}
		if chart == nil {
			chart = tool.NewChart(result.ChartNames())
		}
		var x float64 = -1
		if time {
			x = float64(file.Time)
		}
		chart.Add(x, result.ChartVals())
	}
	return
}

/*
//loadAllGraphData
func loadAllGraphData(time bool, files []*project.File) (tool.GraphData, float64) {
	graphData := make(map[string]tool.GraphData)
	allMax := make(map[string]float64)
	for _, f := range files {
		results, err := db.GraphResults(f.Id)
		if err != nil || results == nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		for _, result := range results {
			if _, ok := graphData[result.GetName()]; !ok {
				graphData[result.GetName()] = make(tool.GraphData, 4)
			}
			allMax[result.GetName()] = result.AddGraphData(
				allMax[result.GetName()], x, graphData[result.GetName()])
		}
	}
	max := 0.0
	for _, v := range allMax {
		max = math.Max(max, v)
	}
	allData := make(tool.GraphData, 0)
	for key, val := range graphData {
		scale := max / allMax[key]
		for _, data := range val {
			if data == nil {
				break
			}
			point := data["data"].([]map[string]float64)
			for _, vals := range point {
				vals["y"] *= scale
			}
			data["data"] = point
			allData = append(allData, data)
		}
	}
	return allData, max
}
*/
