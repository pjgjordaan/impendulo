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
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "JPF"
)

type (
	Result struct {
		Id     bson.ObjectId `bson:"_id"`
		FileId bson.ObjectId `bson:"fileid"`
		Name   string        `bson:"name"`
		Report *Report       `bson:"report"`
		GridFS bool          `bson:"gridfs"`
		Type   string        `bson:"type"`
	}
)

func (r *Result) GetType() string {
	return r.Type
}

//SetReport is used to change r result's report. R comes in handy
//when putting data into/getting data out of GridFS
func (r *Result) SetReport(report result.Reporter) {
	if report == nil {
		r.Report = nil
	} else {
		r.Report = report.(*Report)
	}
}

//OnGridFS determines whether r structure is partially stored on the
//GridFS.
func (r *Result) OnGridFS() bool {
	return r.GridFS
}

func (r *Result) GetTestId() bson.ObjectId {
	return ""
}

//String allows us to print r struct nicely.
func (r *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nReport: %s\n",
		r.Id, r.FileId, r.Name, r.Report)
}

//GetName
func (r *Result) GetName() string {
	return r.Name
}

//GetId
func (r *Result) GetId() bson.ObjectId {
	return r.Id
}

//GetFileId
func (r *Result) GetFileId() bson.ObjectId {
	return r.FileId
}

//Success is true if no errors were found.
func (r *Result) Success() bool {
	return r.Report.Success()
}

//GetReport
func (r *Result) Reporter() result.Reporter {
	return r.Report
}

//ChartVals
func (r *Result) ChartVals() []*result.ChartVal {
	return []*result.ChartVal{
		&result.ChartVal{Name: "Total Errors", Y: float64(r.Report.Total), FileId: r.FileId},
		&result.ChartVal{Name: "Unique Errors", Y: float64(r.Report.ErrorCount()), FileId: r.FileId},
	}
}

func (r *Result) Template() string {
	return "jpfresult"
}

func (r *Result) Lines() []*result.Line {
	return r.Report.Lines()
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
