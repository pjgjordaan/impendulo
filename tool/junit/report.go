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
	"encoding/gob"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"labix.org/v2/mgo/bson"
)

type (
	XML struct {
		Name  string  `xml:"name,attr"`
		Tests int     `xml:"tests,attr"`
		Time  float64 `xml:"time,attr"`
		//Results is all the failed testcases.
		Results TestCases `xml:"testcase"`
	}

	//Report is created from XML output by a Ant JUnit task.
	Report struct {
		Id       bson.ObjectId
		Errors   TestCases
		Failures TestCases
		Name     string
		Tests    int
		Time     float64
	}

	//TestCase represents a failed testcase
	TestCase struct {
		ClassName string   `xml:"classname,attr"`
		Name      string   `xml:"name,attr"`
		Time      float64  `xml:"time,attr"`
		Failure   *Details `xml:"failure"`
		Error     *Details `xml:"error"`
		location  *Location
	}

	//TestCases represents all the testcases which a class failed.
	//It implements sort.Sort
	TestCases []*TestCase

	Details struct {
		Message string `xml:"message,attr"`
		Type    string `xml:"type,attr"`
		Value   string `xml:",innerxml"`
	}
	Location struct {
		Source string
		Line   int
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (*Report, error) {
	var x *XML
	if e := xml.Unmarshal(data, &x); e != nil && x == nil {
		return nil, tool.NewXMLError(e, "junit/junitResult.go")
	}
	return newReport(x, id), nil
}

func newReport(x *XML, id bson.ObjectId) *Report {
	errors := make(TestCases, 0, len(x.Results))
	failures := make(TestCases, 0, len(x.Results))
	for _, tc := range x.Results {
		if e := tc.SetLocation(); e != nil {
			util.Log(e)
		}
		if tc.Error != nil {
			errors = append(errors, tc)
		}
		if tc.Failure != nil {
			failures = append(failures, tc)
		}
	}
	sort.Sort(errors)
	sort.Sort(failures)
	return &Report{Name: x.Name, Time: x.Time, Tests: x.Tests, Id: id, Errors: errors, Failures: failures}
}

//Success
func (r *Report) Success() bool {
	return len(r.Errors) == 0 && len(r.Failures) == 0
}

//String
func (r *Report) String() string {
	return fmt.Sprintf("Id: %q; Success: %t; Tests: %d; Errors: %d; Failures: %d; Name: %s; \n Results: %s",
		r.Id, r.Success(), r.Tests, r.Name, r.Errors, r.Failures)
}

func (r *Report) GetErrors(num int) TestCases {
	if len(r.Errors) < num {
		return r.Errors
	} else {
		return r.Errors[:num]
	}
}

func (r *Report) GetFailures(num int) TestCases {
	if len(r.Failures) < num {
		return r.Failures
	} else {
		return r.Failures[:num]
	}
}

func (r *Report) Passed() int {
	return r.Tests - len(r.Errors) - len(r.Failures)
}

func (r *Report) Lines() []*result.Line {
	lines := make([]*result.Line, 0, len(r.Errors))
	for _, e := range r.Errors {
		if l := e.Location(); l != nil {
			sp := strings.Split(e.Error.Message, ":")
			d := ""
			if len(sp) > 1 {
				d = sp[1]
			}
			lines = append(lines, &result.Line{Title: util.ClassName(sp[0]), Description: d, Start: l.Line, End: l.Line})
		}

	}
	return lines
}

func (t *TestCase) HasData() bool {
	return strings.HasSuffix(t.Name, "txt")

}

//String
func (t *TestCase) String() string {
	return fmt.Sprintf("ClassName: %s; Name: %s; Time: %f; \n Failure: %s\n",
		t.ClassName, t.Name, t.Time, t.Failure)
}

//IsFailure
func (t *TestCase) IsFailure() bool {
	return t.Failure != nil
}

//Len
func (t TestCases) Len() int {
	return len(t)
}

//Swap
func (t TestCases) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

//Less
func (t TestCases) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}

//String
func (t TestCases) String() string {
	s := ""
	for _, c := range t {
		s += c.String()
	}
	return s
}

//String
func (d *Details) String() string {
	return fmt.Sprintf("Message: %s; Type: %s; Value: %s",
		d.Message, d.Type, d.Value)
}

func (t *TestCase) Location() *Location {
	if t.location == nil && t.Error != nil {
		t.location, _ = t.Error.Location()
	}
	return t.location
}

func (t *TestCase) SetLocation() error {
	if t.Error == nil {
		return nil
	}
	l, e := t.Error.Location()
	if e != nil {
		return e
	}
	t.location = l
	return nil
}

func (d *Details) Location() (*Location, error) {
	i := strings.Index(d.Value, "Caused by:")
	if i == -1 {
		return nil, errors.New("\"Caused by:\" not found")
	}
	s := d.Value[i:]
	i = strings.Index(s, "at ")
	if i == -1 {
		return nil, errors.New("\"at \" not found")
	}
	s = s[i:]
	i = strings.Index(s, "(")
	if i == -1 {
		return nil, errors.New("\"(\" not found")
	}
	s = s[i+1:]
	i = strings.Index(s, ":")
	if i == -1 {
		return nil, errors.New("\":\" not found")
	}
	src := s[:i]
	end := strings.Index(s, ")")
	if end == -1 {
		return nil, errors.New("\")\" not found")
	}
	if end <= i+1 {
		return nil, errors.New("line not found")
	}
	l, e := convert.Int(s[i+1 : end])
	if e != nil {
		return nil, e
	}
	return &Location{Source: src, Line: l}, nil
}
