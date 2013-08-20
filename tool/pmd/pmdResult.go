package pmd

import (
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"html/template"
	"labix.org/v2/mgo/bson"
	"math"
)

const NAME = "PMD"

type PMDResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	Time   int64         "time"
	Data   *PMDReport    "data"
}

func (this *PMDResult) GetName() string {
	return this.Name
}

func (this *PMDResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *PMDResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *PMDResult) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d", this.Data.Errors)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *PMDResult) GetData() interface{} {
	return this.Data
}

func (this *PMDResult) Template(current bool) string {
	if current {
		return "pmdCurrent"
	} else {
		return "pmdNext"
	}
}

func (this *PMDResult) Success() bool {
	return true
}

func (this *PMDResult) AddGraphData(max float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("PMD Errors")
	}
	y := float64(this.Data.Errors)
	tool.AddCoords(graphData[0], y)
	return math.Max(max, y)
}

func NewResult(fileId bson.ObjectId, data []byte) (res *PMDResult, err error) {
	res = &PMDResult{Id: bson.NewObjectId(), FileId: fileId, Name: NAME, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *PMDReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		err = tool.NewXMLError(err, "pmd/pmdResult.go")
		return
	}
	res.Id = id
	res.Errors = 0
	for _, f := range res.Files {
		res.Errors += len(f.Violations)
	}
	return
}

func (this *PMDResult) String() (ret string) {
	if this.Data != nil {
		ret = this.Data.String()
	}
	ret += this.Id.Hex()
	return
}

type PMDReport struct {
	Id      bson.ObjectId
	Version string  `xml:"version,attr"`
	Files   []*File `xml:"file"`
	Errors  int
}

func (this *PMDReport) String() (ret string) {
	ret = fmt.Sprintf("Report{ Errors: %d\n.", this.Errors)
	if this.Files != nil {
		ret += "Files: \n"
		for _, f := range this.Files {
			ret += f.String()
		}
	}
	ret += "}\n"
	return
}

type File struct {
	Name       string       `xml:"name,attr"`
	Violations []*Violation `xml:"violation"`
}

func (this *File) String() (ret string) {
	ret = fmt.Sprintf("File{ Name: %s\n.", this.Name)
	if this.Violations != nil {
		ret += "Violations: \n"
		for _, v := range this.Violations {
			ret += v.String()
		}
	}
	ret += "}\n"
	return
}

type Violation struct {
	Begin       int          `xml:"beginline,attr"`
	End         int          `xml:"endline,attr"`
	Rule        string       `xml:"rule,attr"`
	RuleSet     string       `xml:"ruleset,attr"`
	Url         template.URL `xml:"externalInfoUrl,attr"`
	Priority    int          `xml:"priority,attr"`
	Description string       `xml:",innerxml"`
}

func (this *Violation) String() (ret string) {
	ret = fmt.Sprintf("Violation{ Begin: %d; End: %d; Rule: %s; RuleSet: %s; "+
		"Priority: %d; Description: %s}\n",
		this.Begin, this.End, this.Rule, this.RuleSet,
		this.Priority, this.Description)
	return
}
