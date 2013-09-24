//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package webserver

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

type (
	//GraphArgs represents arguments which are passed to rickshaw
	//in order to draw a graph.
	GraphArgs map[string]interface{}
)

//LoadResultGraphData calculates GraphArgs for a given result.
func LoadResultGraphData(result, tipe string, files []*project.File) (graphArgs GraphArgs) {
	var graphData tool.GraphData
	max := -1.0
	time := tipe == "time"
	switch result {
	case "All":
		graphData, max = loadAllGraphData(time, files)
	case tool.CODE:
	case tool.SUMMARY:
	default:
		graphData, max = loadGraphData(result, files, time)
	}
	if max == -1.0 {
		return
	}
	graphArgs = make(GraphArgs)
	graphArgs["max"] = max + max*0.05
	graphArgs["series"] = graphData
	graphArgs["height"] = 400
	graphArgs["width"] = 700
	graphArgs["interpolation"] = "linear"
	graphArgs["renderer"] = "line"
	graphArgs["type"] = tipe
	return
}

func loadGraphData(name string, files []*project.File, time bool) (data tool.GraphData, max float64) {
	max = -1.0
	for _, f := range files {
		file, err := db.File(bson.M{project.ID: f.Id}, bson.M{project.TIME: 1, project.RESULTS: 1})
		if err != nil {
			continue
		}
		result, err := db.GraphResult(name,
			bson.M{project.ID: file.Results[name]}, nil)
		if err != nil {
			continue
		}
		if data == nil {
			data = result.CreateGraphData()
		}
		var x float64 = -1
		if time {
			x = float64(file.Time)
		}
		max = result.AddGraphData(max, x, data)
	}
	return
}

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
