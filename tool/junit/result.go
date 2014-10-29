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

package junit

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	NAME     = "JUnit"
	FAILURES = "Failures"
	ERRORS   = "Errors"
)

type (
	Result struct {
		Id       bson.ObjectId `bson:"_id"`
		FileId   bson.ObjectId `bson:"fileid"`
		TestId   bson.ObjectId `bson:"testid"`
		TestName string        `bson:"name"`
		Report   *Report       `bson:"report"`
		GridFS   bool          `bson:"gridfs"`
		Type     string        `bson:"type"`
	}
)

//SetReport
func (r *Result) SetReport(report result.Reporter) {
	if report == nil {
		r.Report = nil
	} else {
		r.Report = report.(*Report)
	}
}

//OnGridFS
func (r *Result) OnGridFS() bool {
	return r.GridFS
}

//String
func (r *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Report: %s",
		r.Id, r.FileId, r.TestName, r.Report)
}

//GetName
func (r *Result) GetName() string {
	return r.TestName
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
	return r.TestId
}

func (r *Result) Reporter() result.Reporter {
	return r.Report
}

//ChartVals
func (r *Result) ChartVals() []*result.ChartVal {
	return []*result.ChartVal{
		&result.ChartVal{Name: "Failures", Y: float64(len(r.Report.Failures)), FileId: r.FileId},
		&result.ChartVal{Name: "Errors", Y: float64(len(r.Report.Errors)), FileId: r.FileId},
	}
}

func (r *Result) ChartVal(n string) (*result.ChartVal, error) {
	switch n {
	case FAILURES:
		return &result.ChartVal{Name: FAILURES, Y: float64(len(r.Report.Failures)), FileId: r.FileId}, nil
	case ERRORS:
		return &result.ChartVal{Name: ERRORS, Y: float64(len(r.Report.Errors)), FileId: r.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown ChartVal %s", n)
	}
}

func Types() []string {
	return []string{FAILURES, ERRORS}
}

func (r *Result) Template() string {
	return "junitresult"
}

func (r *Result) GetType() string {
	return r.Type
}

func (r *Result) Lines() []*result.Line {
	return r.Report.Lines()
}

//NewResult creates a new junit.Result from provided XML data.
func NewResult(fileId, testId bson.ObjectId, name string, data []byte) (*Result, error) {
	id := bson.NewObjectId()
	r, e := NewReport(id, data)
	if e != nil {
		return nil, e
	}
	return &Result{
		Id:       id,
		FileId:   fileId,
		TestId:   testId,
		TestName: name,
		GridFS:   len(data) > tool.MAX_SIZE,
		Type:     NAME,
		Report:   r,
	}, nil
}
