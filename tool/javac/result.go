//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, r
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, r
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//R SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF R
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package javac

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	NAME     = "Javac"
	ERRORS   = "Errors"
	WARNINGS = "Warnings"
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

//SetReport is used to change r result's report. R comes in handy
//when putting data into/getting data out of GridFS
func (r *Result) SetReport(report result.Reporter) {
	if report == nil {
		r.Report = nil
	} else {
		r.Report = report.(*Report)
	}
}

func (r *Result) GetTestId() bson.ObjectId {
	return ""
}

//OnGridFS
func (r *Result) OnGridFS() bool {
	return r.GridFS
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

func (r *Result) Reporter() result.Reporter {
	return r.Report
}

func (r *Result) Warnings() int {
	if r.Report.Warnings() {
		return r.Report.Count
	}
	return 0
}

func (r *Result) Errors() int {
	if r.Report.Errors() {
		return r.Report.Count
	}
	return 0
}

//Values
func (r *Result) Values() []*result.Value {
	return []*result.Value{
		&result.Value{Name: ERRORS, V: float64(r.Errors()), FileId: r.FileId},
		&result.Value{Name: WARNINGS, V: float64(r.Warnings()), FileId: r.FileId},
	}
}

func (r *Result) Value(n string) (*result.Value, error) {
	switch n {
	case ERRORS:
		return &result.Value{Name: ERRORS, V: float64(r.Errors()), FileId: r.FileId}, nil
	case WARNINGS:
		return &result.Value{Name: WARNINGS, V: float64(r.Warnings()), FileId: r.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown Value %s", n)
	}
}

func (r *Result) Types() []string {
	return []string{ERRORS, WARNINGS}
}

func (r *Result) Template() string {
	return "javacresult"
}

func (r *Result) GetType() string {
	return r.Type
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
		Type:   NAME,
	}
}
