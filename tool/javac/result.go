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

package javac

import (
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "Javac"
)

type (
	//Result is a Javac implementation of tool.ToolResult, tool.DisplayResult
	//and tool.ChartResult. It contains the result of compiling a Java source
	//file using the specified Java compiler.
	Result struct {
		Id     bson.ObjectId "_id"
		FileId bson.ObjectId "fileid"
		Name   string        "name"
		Report *Report       "report"
		GridFS bool          "gridfs"
	}
)

//SetReport is used to change this result's report. This comes in handy
//when putting data into/getting data out of GridFS
func (this *Result) SetReport(report tool.Report) {
	if report == nil {
		this.Report = nil
	} else {
		this.Report = report.(*Report)
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
	if this.Report.Success() {
		body = "Compiled successfully."
	} else {
		body = "No compile."
	}
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//GetReport
func (this *Result) GetReport() tool.Report {
	return this.Report
}

//ChartVals
func (this *Result) ChartVals() []tool.ChartVal {
	var yE, yW int
	if this.Report.Errors() {
		yE = this.Report.Count
	}
	if this.Report.Warnings() {
		yW = this.Report.Count
	}
	return []tool.ChartVal{
		{"Errors", yE, true},
		{"Warnings", yW, false},
	}
}

func (this *Result) Template() string {
	return "javacresult"
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
		Report: NewReport(id, data),
	}
}
