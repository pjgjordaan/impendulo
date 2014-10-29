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
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	NAME   = "Checkstyle"
	ERRORS = "Errors"
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

func (this *Result) GetType() string {
	return this.Type
}

//SetReport is used to change this result's report. This comes in handy
//when putting data into/getting data out of GridFS
func (r *Result) SetReport(report result.Reporter) {
	if report == nil {
		r.Report = nil
	} else {
		r.Report = report.(*Report)
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

func (r *Result) GetTestId() bson.ObjectId {
	return ""
}

func (r *Result) Reporter() result.Reporter {
	return r.Report
}

//ChartVals
func (r *Result) ChartVals() []*result.ChartVal {
	return []*result.ChartVal{
		&result.ChartVal{Name: ERRORS, Y: float64(r.Report.Errors), FileId: r.FileId},
	}
}

func (r *Result) ChartVal(n string) (*result.ChartVal, error) {
	switch n {
	case ERRORS:
		return &result.ChartVal{Name: ERRORS, Y: float64(r.Report.Errors), FileId: r.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown ChartVal %s", n)
	}
}

func Types() []string {
	return []string{ERRORS}
}

func (r *Result) Template() string {
	return "checkstyleresult"
}

func (r *Result) Lines() []*result.Line {
	return r.Report.Lines()
}

//NewResult creates a new Checkstyle Result.
//Any errors returned will be XML errors due to extracting a Report from
//data.
func NewResult(fileId bson.ObjectId, data []byte) (*Result, error) {
	id := bson.NewObjectId()
	r, e := NewReport(id, data)
	if e != nil {
		return nil, e
	}
	return &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		GridFS: len(data) > tool.MAX_SIZE,
		Type:   NAME,
		Report: r,
	}, nil
}
