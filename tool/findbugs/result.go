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

package findbugs

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "Findbugs"
	ALL  = "All"
	P1   = "Priority 1"
	P2   = "Priority 2"
	P3   = "Priority 3"
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
func (r *Result) OnGridFS() bool {
	return r.GridFS
}

//String
func (r *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Report: %s", r.Id, r.FileId, r.Name, r.Report)
}

//GetName
func (r *Result) GetName() string {
	return r.Name
}

//GetId
func (r *Result) GetId() bson.ObjectId {
	return r.Id
}

func (this *Result) GetTestId() bson.ObjectId {
	return ""
}

//GetFileId
func (r *Result) GetFileId() bson.ObjectId {
	return r.FileId
}

func (r *Result) Reporter() result.Reporter {
	return r.Report
}

func (r *Result) Lines() []*result.Line {
	return r.Report.Lines()
}

//Values
func (r *Result) Values() []*result.Value {
	return []*result.Value{
		&result.Value{Name: "All", V: float64(r.Report.Summary.BugCount), FileId: r.FileId},
		&result.Value{Name: "Priority 1", V: float64(r.Report.Summary.Priority1), FileId: r.FileId},
		&result.Value{Name: "Priority 2", V: float64(r.Report.Summary.Priority2), FileId: r.FileId},
		&result.Value{Name: "Priority 3", V: float64(r.Report.Summary.Priority3), FileId: r.FileId},
	}
}

func (r *Result) Value(n string) (*result.Value, error) {
	switch n {
	case ALL:
		return &result.Value{Name: ALL, V: float64(r.Report.Summary.BugCount), FileId: r.FileId}, nil
	case P1:
		return &result.Value{Name: P1, V: float64(r.Report.Summary.Priority1), FileId: r.FileId}, nil
	case P2:
		return &result.Value{Name: P2, V: float64(r.Report.Summary.Priority2), FileId: r.FileId}, nil
	case P3:
		return &result.Value{Name: P3, V: float64(r.Report.Summary.Priority3), FileId: r.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown Value %s", n)
	}
}

func (r *Result) Types() []string {
	return []string{ALL, P1, P2, P3}
}

func (r *Result) Template() string {
	return "findbugsresult"
}

func (r *Result) GetType() string {
	return r.Type
}

//NewResult
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
