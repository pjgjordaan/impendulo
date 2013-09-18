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
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
	"math"
)

type (
	//GraphArgs represents arguments which are passed to rickshaw
	//in order to draw a graph.
	GraphArgs map[string]interface{}

	//GraphData represents the x and y values used to draw the graphs.
	GraphData []map[string]interface{}
)

//LoadResultGraphData calculates GraphArgs for a given result.
func LoadResultGraphData(result, tipe string, files []*project.File) (graphArgs GraphArgs) {
	var graphData GraphData
	max := -1.0
	time := tipe == "time"
	switch result {
	case javac.NAME:
		graphData, max = loadJavacGraphData(time, files)
	case jpf.NAME:
		graphData, max = loadJPFGraphData(time, files)
	case findbugs.NAME:
		graphData, max = loadFindbugsGraphData(time, files)
	case pmd.NAME:
		graphData, max = loadPMDGraphData(time, files)
	case checkstyle.NAME:
		graphData, max = loadCheckstyleGraphData(time, files)
	case "All":
		graphData, max = loadAllGraphData(time, files)
	case tool.CODE:
	case tool.SUMMARY:
	default:
		graphData, max = loadJUnitGraphData(result, time, files)
	}
	if max == -1.0 {
		return
	}
	graphArgs = make(GraphArgs)
	graphArgs["max"] = max + max*0.05
	graphArgs["series"] = graphData
	/*	if result == javac.NAME {
		graphArgs["yformat"] = map[string]interface{}{"0": "Failure", "1": "Success"}
		graphArgs["min"] = -0.05
	}*/
	graphArgs["height"] = 400
	graphArgs["width"] = 700
	graphArgs["interpolation"] = "linear"
	graphArgs["renderer"] = "line"
	graphArgs["type"] = tipe
	return
}

//loadJavacGraphData
func loadJavacGraphData(time bool, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 2)
	max := -1.0
	for _, f := range files {
		result, err := db.JavacResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		max = result.AddGraphData(max, x, graphData)
	}
	return graphData, max
}

//loadJUnitGraphData
func loadJUnitGraphData(testName string, time bool, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 3)
	max := -1.0
	for _, f := range files {
		result, err := db.JUnitResult(
			bson.M{project.FILEID: f.Id, project.NAME: testName}, nil)
		if err != nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		max = result.AddGraphData(max, x, graphData)
	}
	return graphData, max
}

//loadJPFGraphData
func loadJPFGraphData(time bool, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 1)
	max := -1.0
	for _, f := range files {
		result, err := db.JPFResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		max = result.AddGraphData(max, x, graphData)
	}
	return graphData, max
}

//loadFindbugsGraphData
func loadFindbugsGraphData(time bool, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 4)
	max := -1.0
	for _, f := range files {
		result, err := db.FindbugsResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		max = result.AddGraphData(max, x, graphData)
	}
	return graphData, max
}

//loadCheckstyleGraphData
func loadCheckstyleGraphData(time bool, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 1)
	max := -1.0
	for _, f := range files {
		result, err := db.CheckstyleResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		max = result.AddGraphData(max, x, graphData)
	}
	return graphData, max
}

//loadPMDGraphData
func loadPMDGraphData(time bool, files []*project.File) (GraphData, float64) {
	graphData := make(GraphData, 1)
	max := -1.0
	for _, f := range files {
		result, err := db.PMDResult(
			bson.M{project.FILEID: f.Id}, nil)
		if err != nil || result.Data == nil {
			continue
		}
		var x float64 = -1
		if time {
			x = float64(f.Time)
		}
		max = result.AddGraphData(max, x, graphData)
	}
	return graphData, max
}

//loadAllGraphData
func loadAllGraphData(time bool, files []*project.File) (GraphData, float64) {
	graphData := make(map[string]GraphData)
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
				graphData[result.GetName()] = make(GraphData, 4)
			}
			allMax[result.GetName()] = result.AddGraphData(
				allMax[result.GetName()], x, graphData[result.GetName()])
		}
	}
	max := 0.0
	for _, v := range allMax {
		max = math.Max(max, v)
	}
	allData := make(GraphData, 0)
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
