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

package junit

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "JUnit Test"
)

type (
	//Result is an implementation of ToolResult, DisplayResult
	//and GraphResult for JUnit test results.
	Result struct {
		Id       bson.ObjectId "_id"
		FileId   bson.ObjectId "fileid"
		TestName string        "name"
		Data     *Report       "data"
		GridFS   bool          "gridfs"
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

//String
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Data: %s",
		this.Id, this.FileId, this.TestName, this.Data)
}

//GetName
func (this *Result) GetName() string {
	return this.TestName
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
	body := fmt.Sprintf("Tests: %d \n Failures: %d \n Errors: %d \n Time: %f",
		this.Data.Tests, this.Data.Failures, this.Data.Errors, this.Data.Time)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//GetData
func (this *Result) GetData() interface{} {
	return this.Data
}

//Success
func (this *Result) Success() bool {
	return this.Data.Success()
}

//CreateGraphData
func (this *Result) CreateGraphData() (graphData tool.GraphData) {
	graphData = make(tool.GraphData, 3)
	graphData[0] = tool.CreateChart(this.GetName() + " Errors")
	graphData[1] = tool.CreateChart(this.GetName() + " Failures")
	graphData[2] = tool.CreateChart(this.GetName() + " Successes")
	return
}

//AddGraphData
func (this *Result) AddGraphData(max, x float64, graphData tool.GraphData) float64 {
	yE := float64(this.Data.Errors)
	yF := float64(this.Data.Failures)
	yS := float64(this.Data.Tests - this.Data.Failures - this.Data.Errors)
	tool.AddCoords(graphData[0], x, yE)
	tool.AddCoords(graphData[1], x, yF)
	tool.AddCoords(graphData[2], x, yS)
	return math.Max(max, math.Max(yE, math.Max(yF, yS)))
}

//NewResult creates a new junit.Result from provided XML data.
func NewResult(fileId bson.ObjectId, name string, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:       bson.NewObjectId(),
		FileId:   fileId,
		TestName: name,
		GridFS:   gridFS,
	}
	res.Data, err = NewReport(res.Id, data)
	return
}
