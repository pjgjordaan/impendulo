//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package junit

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"sort"
	"strings"
)

type (
	//Report is created from XML output by a Ant JUnit task.
	Report struct {
		Id bson.ObjectId
		//Errors is the number of runtime exceptions
		//which occurred when running the test cases.
		Errors int `xml:"errors,attr"`
		//Failures is the number of test cases which produced
		//an invalid result.
		Failures int     `xml:"failures,attr"`
		Name     string  `xml:"name,attr"`
		Tests    int     `xml:"tests,attr"`
		Time     float64 `xml:"time,attr"`
		//Results is all the failed testcases.
		Results TestCases `xml:"testcase"`
	}

	//TestCase represents a failed testcase
	TestCase struct {
		ClassName string   `xml:"classname,attr"`
		Name      string   `xml:"name,attr"`
		Time      float64  `xml:"time,attr"`
		Fail      *Failure `xml:"failure"`
	}

	//TestCases represents all the testcases which a class failed.
	//It implements sort.Sort
	TestCases []*TestCase

	//Failure gives details as to why a test case failed.
	Failure struct {
		Message string `xml:"message,attr"`
		Type    string `xml:"type,attr"`
		Value   string `xml:",innerxml"`
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (res *Report, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		if res == nil {
			err = tool.NewXMLError(err, "junit/junitResult.go")
			return
		} else {
			err = nil
		}
	}
	sort.Sort(res.Results)
	res.Id = id
	return
}

//Success
func (this *Report) Success() bool {
	return this.Errors == 0 && this.Failures == 0
}

//String
func (this *Report) String() string {
	return fmt.Sprintf("Id: %q; Success: %t; Tests: %d; Errors: %d; Failures: %d; Name: %s; \n Results: %s",
		this.Id, this.Success, this.Tests, this.Errors, this.Failures, this.Name, this.Results)
}

//GetResults
func (this *Report) GetResults(num int) TestCases {
	if len(this.Results) < num {
		return this.Results
	} else {
		return this.Results[:num]
	}
}

//String
func (this *TestCase) String() string {
	return fmt.Sprintf("ClassName: %s; Name: %s; Time: %f; \n Failure: %s\n",
		this.ClassName, this.Name, this.Time, this.Fail)
}

//IsFailure
func (this *TestCase) IsFailure() bool {
	return this.Fail != nil && len(strings.TrimSpace(this.Fail.Type)) > 0
}

//Len
func (this TestCases) Len() int {
	return len(this)
}

//Swap
func (this TestCases) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

//Less
func (this TestCases) Less(i, j int) bool {
	return this[i].Name < this[j].Name
}

//String
func (this TestCases) String() (ret string) {
	for _, t := range this {
		ret += t.String()
	}
	return ret
}

//String
func (this *Failure) String() string {
	return fmt.Sprintf("Message: %s; Type: %s; Value: %s",
		this.Message, this.Type, this.Value)
}
