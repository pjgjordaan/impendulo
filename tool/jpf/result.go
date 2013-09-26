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

package jpf

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "JPF"
)

type (
	//Result is a JPF implementation of tool.ToolResult, tool.DisplayResult and tool.ChartResult.
	Result struct {
		Id     bson.ObjectId "_id"
		FileId bson.ObjectId "fileid"
		Name   string        "name"
		Data   *Report       "data"
		GridFS bool          "gridfs"
	}
)

//SetData is used to change this result's data. This comes in handy
//when putting data into/getting data out of GridFS
func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

//OnGridFS determines whether this structure is partially stored on the
//GridFS.
func (this *Result) OnGridFS() bool {
	return this.GridFS
}

//String allows us to print this struct nicely.
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nData: %s\n",
		this.Id, this.FileId, this.Name, this.Data)
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

//Summary is the errors found by JPF.
func (this *Result) Summary() *tool.Summary {
	body := fmt.Sprintf("Result: %s \n Errors: %d",
		this.Data.Findings.Description, len(this.Data.Findings.Errors))
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//Success is true if no errors were found.
func (this *Result) Success() bool {
	return len(this.Data.Findings.Errors) == 0
}

//GetData
func (this *Result) GetData() interface{} {
	return this.Data
}

//ChartNames
func (this *Result) ChartNames() []string {
	return []string{
		"Error Detection Time",
	}
}

//ChartVals
func (this *Result) ChartVals() map[string]float64 {
	var yT float64
	if this.Data.Errors() == 0 {
		yT = 0
	} else {
		yT = float64(this.Data.Stats.Time)
	}
	return map[string]float64{
		"Error Detection Time": yT,
	}
}

//NewResult creates a new JPF result. The data []byte is in XML format and
//therefore allows us to generate a JPF report from it.
func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	res = &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		GridFS: len(data) > tool.MAX_SIZE,
	}
	res.Data, err = NewReport(res.Id, data)
	return
}
