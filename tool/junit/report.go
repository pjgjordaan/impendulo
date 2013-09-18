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
		Id       bson.ObjectId
		Errors   int       `xml:"errors,attr"`
		Failures int       `xml:"failures,attr"`
		Name     string    `xml:"name,attr"`
		Tests    int       `xml:"tests,attr"`
		Time     float64   `xml:"time,attr"`
		Results  TestCases `xml:"testcase"`
	}

	TestCase struct {
		ClassName string   `xml:"classname,attr"`
		Name      string   `xml:"name,attr"`
		Time      float64  `xml:"time,attr"`
		Fail      *Failure `xml:"failure"`
	}

	TestCases []*TestCase

	Failure struct {
		Message string `xml:"message,attr"`
		Type    string `xml:"type,attr"`
		Value   string `xml:",innerxml"`
	}
)

func init() {
	gob.Register(new(Report))
}

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

func (this *Report) Success() bool {
	return this.Errors == 0 && this.Failures == 0
}

func (this *Report) String() string {
	return fmt.Sprintf("Id: %q; Success: %t; Tests: %d; Errors: %d; Failures: %d; Name: %s; \n Results: %s",
		this.Id, this.Success, this.Tests, this.Errors, this.Failures, this.Name, this.Results)
}

func (this *Report) GetResults(num int) TestCases {
	if len(this.Results) < num {
		return this.Results
	} else {
		return this.Results[:num]
	}
}

func (this *TestCase) String() string {
	return fmt.Sprintf("ClassName: %s; Name: %s; Time: %f; \n Failure: %s\n",
		this.ClassName, this.Name, this.Time, this.Fail)
}

func (this *TestCase) IsFailure() bool {
	return this.Fail != nil && len(strings.TrimSpace(this.Fail.Type)) > 0
}

func (this TestCases) Len() int {
	return len(this)
}

func (this TestCases) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this TestCases) Less(i, j int) bool {
	return this[i].Name < this[j].Name
}

func (this TestCases) String() (ret string) {
	for _, t := range this {
		ret += t.String()
	}
	return ret
}

func (this *Failure) String() string {
	return fmt.Sprintf("Message: %s; Type: %s; Value: %s",
		this.Message, this.Type, this.Value)
}
