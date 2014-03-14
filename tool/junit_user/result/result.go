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

package result

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "JUnit_User"
)

type (
	//Result is an implementation of ToolResult, DisplayResult
	//and ChartResult for JUnit user test results.
	Result struct {
		Id         bson.ObjectId "_id"
		FileId     bson.ObjectId "fileid"
		TestFileId bson.ObjectId "testfileid"
		TestName   string        "name"
		Report     *junit.Report "report"
		GridFS     bool          "gridfs"
		Type       string        "type"
	}
	Report struct {
		TestFileId bson.ObjectId
		*junit.Report
	}
)

func (this *Report) GetTestFileId() bson.ObjectId {
	return this.TestFileId
}

//SetReport
func (this *Result) SetReport(report tool.Report) {
	if report == nil {
		this.Report = nil
	} else {
		this.Report = report.(*junit.Report)
	}
}

//OnGridFS
func (this *Result) OnGridFS() bool {
	return this.GridFS
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
		this.Report.Tests, this.Report.Failures, this.Report.Errors, this.Report.Time)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//GetReport
func (this *Result) GetReport() tool.Report {
	return &Report{this.TestFileId, this.Report}
}

//Success
func (this *Result) Success() bool {
	return this.Report != nil && this.Report.Success()
}

//ChartVals
func (this *Result) ChartVals() []*tool.ChartVal {
	return []*tool.ChartVal{
		&tool.ChartVal{"Errors", float64(this.Report.Errors), true, this.FileId},
		&tool.ChartVal{"Failures", float64(this.Report.Failures), true, this.FileId},
	}
}

func (this *Result) Template() string {
	return "junituserresult"
}

func (this *Result) GetType() string {
	return this.Type
}

func New(testFileId bson.ObjectId, junitRes tool.ToolResult) (res *Result, err error) {
	res = &Result{
		Id:         bson.NewObjectId(),
		FileId:     junitRes.GetFileId(),
		TestFileId: testFileId,
		TestName:   junitRes.GetName(),
		GridFS:     junitRes.OnGridFS(),
		Report:     junitRes.GetReport().(*junit.Report),
		Type:       NAME,
	}
	res.Report.Id = res.Id
	return
}
