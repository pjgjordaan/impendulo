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

package pmd

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "PMD"
)

type (
	//Result is a PMD implementation of tool.ToolResult,
	//tool.DisplayResult and tool.GraphResult.
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

//Summary consists of the number of errors found by PMD.
func (this *Result) Summary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d", this.Data.Errors)
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
	return true
}

//AddGraphData adds the number of errors found by PMD.
func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("PMD Errors")
	}
	y := float64(this.Data.Errors)
	tool.AddCoords(graphData[0], x, y)
	return math.Max(max, y)
}

//String
func (this *Result) String() (ret string) {
	if this.Data != nil {
		ret = this.Data.String()
	}
	ret += this.Id.Hex()
	return
}

//NewResult creates a new PMD Result.
//Any error returned will be as a result of creating a PMD Report
//from the XML in data.
func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
	}
	res.Data, err = NewReport(res.Id, data)
	return
}
