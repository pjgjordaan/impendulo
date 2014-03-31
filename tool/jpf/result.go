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
		Report *Report       "report"
		GridFS bool          "gridfs"
		Type   string        "type"
	}
)

func (this *Result) GetType() string {
	return this.Type
}

//SetReport is used to change this result's report. This comes in handy
//when putting data into/getting data out of GridFS
func (this *Result) SetReport(report tool.Report) {
	if report == nil {
		this.Report = nil
	} else {
		this.Report = report.(*Report)
	}
}

//OnGridFS determines whether this structure is partially stored on the
//GridFS.
func (this *Result) OnGridFS() bool {
	return this.GridFS
}

//String allows us to print this struct nicely.
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nReport: %s\n",
		this.Id, this.FileId, this.Name, this.Report)
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
	body := fmt.Sprintf("Errors: %d",
		this.Report.ErrorCount())
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//Success is true if no errors were found.
func (this *Result) Success() bool {
	return this.Report.Success()
}

//GetReport
func (this *Result) GetReport() tool.Report {
	return this.Report
}

//ChartVals
func (this *Result) ChartVals() []*tool.ChartVal {
	return []*tool.ChartVal{
		&tool.ChartVal{"Total Errors", float64(this.Report.Total), this.FileId},
		&tool.ChartVal{"Unique Errors", float64(this.Report.ErrorCount()), this.FileId},
	}
}

func (this *Result) Template() string {
	return "jpfresult"
}

//NewResult creates a new JPF result. The data []byte is in XML format and
//therefore allows us to generate a JPF report from it.
func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	id := bson.NewObjectId()
	report, err := NewReport(id, data)
	if err != nil {
		return
	}
	res = &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		GridFS: len(data) > tool.MAX_SIZE,
		Type:   NAME,
		Report: report,
	}
	return
}
