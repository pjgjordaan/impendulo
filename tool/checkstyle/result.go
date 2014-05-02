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

package checkstyle

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "Checkstyle"
)

type (
	//Result is Checkstyle's implementation of tool.ToolResult, tool.DisplayResult
	//and tool.ChartResult
	Result struct {
		Id     bson.ObjectId `bson:"_id"`
		FileId bson.ObjectId `bson:"fileid"`
		Name   string        `bson:"name"`
		Report *Report       `bson:"report"`
		GridFS bool          `bson:"gridfs"`
		Type   string        `bson:"type"`
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

//OnGridFS
func (this *Result) OnGridFS() bool {
	return this.GridFS
}

//String
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nReport: %s\n",
		this.Id, this.FileId, this.Name, this.Report.String())
}

//GetName
func (this *Result) GetName() string {
	return this.Name
}

//Summary is the number of errors Checkstyle found.
func (this *Result) Summary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d",
		this.Report.Errors)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//GetId
func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

//GetFileId
func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

//GetReport
func (this *Result) GetReport() tool.Report {
	return this.Report
}

//Success
func (this *Result) Success() bool {
	return true
}

//ChartVals
func (this *Result) ChartVals() []*tool.ChartVal {
	return []*tool.ChartVal{
		&tool.ChartVal{Name: "Errors", Y: float64(this.Report.Errors), FileId: this.FileId},
	}
}

func (this *Result) Template() string {
	return "checkstyleresult"
}

//NewResult creates a new Checkstyle Result.
//Any errors returned will be XML errors due to extracting a Report from
//data.
func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	id := bson.NewObjectId()
	report, err := NewReport(id, data)
	if err != nil {
		return
	}
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
		Type:   NAME,
		Report: report,
	}
	return
}
