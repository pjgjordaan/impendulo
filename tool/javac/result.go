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

package javac

import (
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "Javac"
)

type (
	//Result is a Javac implementation of tool.ToolResult, tool.DisplayResult
	//and tool.GraphResult. It contains the result of compiling a Java source
	//file using the specified Java compiler.
	Result struct {
		Id     bson.ObjectId "_id"
		FileId bson.ObjectId "fileid"
		Name   string        "name"
		Data   *Report       "data"
		GridFS bool          "gridfs"
	}
)

//SetData
func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

//OnGridFS
func (this *Result) OnGridFS() bool {
	return this.GridFS
}

//GetName
func (this *Result) GetName() string {
	return this.Name
}

//GetId
func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

//GetFileId
func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

//Summary
func (this *Result) Summary() *tool.Summary {
	var body string
	if this.Data.Success() {
		body = "Compiled successfully."
	} else {
		body = "No compile."
	}
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//GetData
func (this *Result) GetData() interface{} {
	return this.Data
}

//AddGraphData
func (this *Result) AddGraphData(max, x float64, graphData tool.GraphData) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart(this.GetName() + " Errors")
		graphData[1] = tool.CreateChart(this.GetName() + " Warnings")
	}
	yE, yW := 0.0, 0.0
	if this.Data.Errors() {
		yE = float64(this.Data.Count)
	} else if this.Data.Warnings() {
		yW = float64(this.Data.Count)
	}
	tool.AddCoords(graphData[0], x, yE)
	tool.AddCoords(graphData[1], x, yW)
	return math.Max(max, math.Max(yE, yW))
}

//CreateGraphData
func (this *Result) CreateGraphData() (graphData tool.GraphData) {
	graphData = make(tool.GraphData, 2)
	graphData[0] = tool.CreateChart(this.GetName() + " Errors")
	graphData[1] = tool.CreateChart(this.GetName() + " Warnings")
	return
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *Result {
	gridFS := len(data) > tool.MAX_SIZE
	id := bson.NewObjectId()
	return &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
		Data:   NewReport(id, data),
	}
}
