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
	"labix.org/v2/mgo/bson"
	"strconv"
)

const (
	NAME = "Findbugs"
)

type (
	//Result is a tool.ToolResult and a tool.DisplayResult.
	//It is used to store the output of running Findbugs.
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

//String
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Report: %s",
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

//Summary
func (this *Result) Summary() *tool.Summary {
	body := fmt.Sprintf("Bugs: %d", this.Report.Summary.BugCount)
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
func (this *Result) ChartVals(summary bool) []tool.ChartVal {
	if summary {
		return []tool.ChartVal{
			{"Findbugs Bugs", float64(this.Report.Summary.BugCount), true},
		}
	} else {
		return []tool.ChartVal{
			{"All", float64(this.Report.Summary.BugCount), true},
			{"Priority 1", float64(this.Report.Summary.Priority1), false},
			{"Priority 2", float64(this.Report.Summary.Priority2), false},
			{"Priority 3", float64(this.Report.Summary.Priority3), false},
		}
	}
}

func (this *Result) Template() string {
	return "findbugsResult"
}

func (this *Result) Bug(id string, index int) (bug *tool.Bug, err error) {
	if index < 0 || index > len(this.Report.Instances) {
		err = fmt.Errorf("Index %d out of bounds for Findbugs Bugs array.", index)
		return
	}
	instance := this.Report.Instances[index]
	if bId := instance.Id.Hex(); bId != id {
		err = fmt.Errorf("Provided id %s does not Findbugs bug id %s.", id, bId)
		return
	}
	content := []interface{}{
		this.Report.PatternMap[instance.Type].Description,
		this.Report.CategoryMap[instance.Category].Description,
		"Priority: " + strconv.Itoa(instance.Priority),
		"Rank: " + strconv.Itoa(instance.Rank),
	}
	bug = tool.NewBug(this, id, content, instance.Line.Start, instance.Line.End)
	return
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
	}
	res.Report, err = NewReport(res.Id, data)
	return
}
