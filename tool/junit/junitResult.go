package junit

import (
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
	"strings"
	"sort"
)

const NAME = "JUnit Test"

type JUnitResult struct {
	Id       bson.ObjectId "_id"
	FileId   bson.ObjectId "fileid"
	TestName string        "name"
	Data     *TestSuite    "data"
}

func (this *JUnitResult) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Data: %s", 
		this.Id, this.FileId, this.TestName, this.Data)
} 

func (this *JUnitResult) GetName() string {
	return this.TestName
}

func (this *JUnitResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *JUnitResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *JUnitResult) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Tests: %d \n Failures: %d \n Errors: %d \n Time: %f",
		this.Data.Tests, this.Data.Failures, this.Data.Errors, this.Data.Time)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *JUnitResult) GetData() interface{} {
	return this.Data
}

func (this *JUnitResult) Template(current bool) string {
	if current {
		return "junitCurrent"
	} else {
		return "junitNext"
	}
}

func (this *JUnitResult) Success() bool {
	return this.Data.Success
}

func (this *JUnitResult) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart(this.GetName() + " Errors")
		graphData[1] = tool.CreateChart(this.GetName() + " Failures")
		graphData[2] = tool.CreateChart(this.GetName() + " Successes")
	}
	yE := float64(this.Data.Errors)
	yF := float64(this.Data.Failures)
	yS := float64(this.Data.Tests - this.Data.Failures - this.Data.Errors)
	tool.AddCoords(graphData[0], x, yE)
	tool.AddCoords(graphData[1], x, yF)
	tool.AddCoords(graphData[2], x, yS)
	return math.Max(max, math.Max(yE, math.Max(yF, yS)))
}



func NewResult(fileId bson.ObjectId, name string, data []byte) (res *JUnitResult, err error) {
	res = &JUnitResult{
		Id: bson.NewObjectId(), 
		FileId: fileId, 
		TestName: name,
	}
	res.Data, err = genReport(res.Id, data)
	return
}

type TestSuite struct {
	Id       bson.ObjectId
	Success  bool
	Errors   int        `xml:"errors,attr"`
	Failures int        `xml:"failures,attr"`
	Name     string     `xml:"name,attr"`
	Tests    int        `xml:"tests,attr"`
	Time     float64    `xml:"time,attr"`
	Results  TestCases `xml:"testcase"`
}


func (this *TestSuite) String() string {
	return fmt.Sprintf("Id: %q; Success: %t; Tests: %d; Errors: %d; Failures: %d; Name: %s; \n Results: %s", 
		this.Id, this.Success, this.Tests, this.Errors, this.Failures, this.Name, this.Results)
} 

func (this *TestSuite) GetResults(num int) TestCases{
	if len(this.Results) < num{
		return this.Results
	} else{
		return this.Results[:num]
	}
}

type TestCase struct {
	ClassName string  `xml:"classname,attr"`
	Name      string  `xml:"name,attr"`
	Time      float64 `xml:"time,attr"`
	Fail      *Failure `xml:"failure"`
}


func (this *TestCase) String() string {
	return fmt.Sprintf("ClassName: %s; Name: %s; Time: %f; \n Failure: %s\n", 
		this.ClassName, this.Name, this.Time, this.Fail)
} 

func (this *TestCase) IsFailure() bool{
	return this.Fail != nil && len(strings.TrimSpace(this.Fail.Type)) > 0
}

type TestCases []*TestCase

func (this TestCases) Len()int{
	return len(this)
} 

func (this TestCases) Swap(i, j int){
	this[i], this[j] = this[j], this[i]
} 

func (this TestCases) Less(i, j int) bool{
	return this[i].Name < this[j].Name
} 


func (this TestCases) String() (ret string) {
	for _, t := range this{
		ret += t.String()
	}
	return ret
} 

type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Value   string `xml:",innerxml"`
}


func (this *Failure) String() string {
	return fmt.Sprintf("Message: %s; Type: %s; Value: %s", 
		this.Message, this.Type, this.Value)
} 

func genReport(id bson.ObjectId, data []byte) (res *TestSuite, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		if res == nil {
			err = tool.NewXMLError(err, "junit/junitResult.go")
			return
		} else {
			err = nil
		}
	}
	sort.Sort(res.Results)
	if res.Errors == 0 && res.Failures == 0 {
		res.Success = true
	} else {
		res.Success = false
	}
	res.Id = id
	return
}
