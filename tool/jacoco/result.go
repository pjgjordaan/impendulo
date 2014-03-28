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

package jacoco

import (
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Result is an implementation of ToolResult, DisplayResult
	//and ChartResult for Jacoco coverage results.
	Result struct {
		Id       bson.ObjectId "_id"
		FileId   bson.ObjectId "fileid"
		TestName string        "name"
		Report   *Report       "report"
		GridFS   bool          "gridfs"
		Type     string        "type"
	}
)

//SetReport
func (r *Result) SetReport(n tool.Report) {
	if n == nil {
		r.Report = nil
	} else {
		r.Report = n.(*Report)
	}
}

//OnGridFS
func (r *Result) OnGridFS() bool {
	return r.GridFS
}

//GetName
func (r *Result) GetName() string {
	return NAME
}

//GetId
func (r *Result) GetId() bson.ObjectId {
	return r.Id
}

//GetFileId
func (r *Result) GetFileId() bson.ObjectId {
	return r.FileId
}

//Summary
func (r *Result) Summary() *tool.Summary {
	return &tool.Summary{
		Name: r.GetName(),
	}
}

//GetReport
func (r *Result) GetReport() tool.Report {
	return r.Report
}

func (r *Result) Success() bool {
	return r.Report != nil
}

func (r *Result) Template() string {
	return "jacocoresult"
}

func (r *Result) GetType() string {
	return r.Type
}

//ChartVals
func (r *Result) ChartVals() []*tool.ChartVal {
	v := make([]*tool.ChartVal, len(r.Report.Counters))
	for i, c := range r.Report.Counters {
		p := util.Round(float64(c.Covered)/float64(c.Covered+c.Missed)*100.0, 2)
		v[i] = &tool.ChartVal{util.Title(c.Type) + " Coverage", p, i == 0, r.FileId}
	}
	return v
}

func NewResult(fileId bson.ObjectId, name string, xml, html []byte) (*Result, error) {
	id := bson.NewObjectId()
	r, e := NewReport(id, xml, html)
	if e != nil {
		return nil, e
	}
	return &Result{
		Id:       id,
		FileId:   fileId,
		TestName: name,
		GridFS:   len(xml)+len(html) > tool.MAX_SIZE,
		Report:   r,
		Type:     NAME,
	}, nil
}
