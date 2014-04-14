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

package db

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"

	"html/template"

	"labix.org/v2/mgo/bson"

	"reflect"
	"testing"
)

func TestResultNames(t *testing.T) {
	s := project.Submission{Id: bson.NewObjectId(), ProjectId: bson.NewObjectId()}
	if e := Add(SUBMISSIONS, s); e != nil {
		t.Error(e)
	}
	f0 := &project.File{Id: bson.NewObjectId(), Name: "Test", SubId: s.Id, Results: bson.M{"resulta-abb": "asda", "resulta-asb": "asda", "result-abb": "asda"}}
	if e := Add(FILES, f0); e != nil {
		t.Error(e)
	}
	f1 := &project.File{Id: bson.NewObjectId(), Name: "Test2", SubId: s.Id, Results: bson.M{"a": "b", "c": "d", "e": "f", "g": "h", "i": "j"}}
	if e := Add(FILES, f1); e != nil {
		t.Error(e)
	}
	f2 := &project.File{Id: bson.NewObjectId(), Name: "Test", SubId: s.Id, Results: bson.M{"wresulta-abb": "asda", "kresulta-asb": "asda", "tresult-abb": "asda"}}
	if e := Add(FILES, f2); e != nil {
		t.Error(e)
	}
	if _, e := ResultNames(s.Id, f0.Name); e != nil {
		t.Error(e)
	}
}

func TestResult(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	file, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = Add(FILES, file)
	if err != nil {
		t.Error(err)
	}
	res := javac.NewResult(file.Id, fileData)
	err = AddResult(res, res.GetName())
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": res.GetId()}
	dbRes, err := ToolResult(matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(res, dbRes) {
		t.Error("Results not equivalent")
	}
}

func javacResult(fileId bson.ObjectId, gridFS bool) *javac.Result {
	id := bson.NewObjectId()
	return &javac.Result{
		Id:     id,
		FileId: fileId,
		Name:   javac.NAME,
		Report: javacReport(id),
		GridFS: gridFS,
		Type:   javac.NAME,
	}
}

func checkstyleResult(fileId bson.ObjectId, gridFS bool) *checkstyle.Result {
	id := bson.NewObjectId()
	return &checkstyle.Result{
		Id:     id,
		FileId: fileId,
		Name:   checkstyle.NAME,
		Report: checkstyleReport(id),
		GridFS: gridFS,
		Type:   checkstyle.NAME,
	}
}

func findbugsResult(fileId bson.ObjectId, gridFS bool) *findbugs.Result {
	id := bson.NewObjectId()
	return &findbugs.Result{
		Id:     id,
		FileId: fileId,
		Name:   findbugs.NAME,
		Report: findbugsReport(id),
		GridFS: gridFS,
		Type:   findbugs.NAME,
	}
}

func report(name string, id bson.ObjectId) interface{} {
	switch name {
	case checkstyle.NAME:
		return checkstyleReport(id)
	case javac.NAME:
		return javacReport(id)
	case findbugs.NAME:
		return findbugsReport(id)
	default:
		return nil
	}
}

func findbugsReport(id bson.ObjectId) *findbugs.Report {
	report := new(findbugs.Report)
	report.Id = id
	report.Instances = []*findbugs.BugInstance{
		{
			Type:     "some bug",
			Priority: 1,
			Rank:     2,
		},
	}
	return report
}

func javacReport(id bson.ObjectId) *javac.Report {
	return &javac.Report{
		Id:    id,
		Type:  tool.ERRORS,
		Count: 4,
		Data:  []byte("some errors were found"),
	}
}

func checkstyleReport(id bson.ObjectId) *checkstyle.Report {
	return &checkstyle.Report{
		Id:      id,
		Version: "1.0",
		Errors:  0,
		Files: []*checkstyle.File{
			{
				Name: "Dummy.java",
				Errors: []*checkstyle.Error{
					{
						Line:     0,
						Column:   1,
						Severity: "Bad",
						Message:  template.HTML("<h2>Bad stuff</h2>"),
						Source:   "Some place.",
					},
				},
			},
		},
	}
}
