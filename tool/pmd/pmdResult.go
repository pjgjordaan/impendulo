package pmd

import (
	"encoding/xml"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/tool"
	"fmt"
	"html/template"
	"labix.org/v2/mgo/bson"
)

const NAME = "PMD"

type PMDResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data   *PMDReport    "data"
}

func (this *PMDResult) GetName() string {
	return NAME
}

func (this *PMDResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *PMDResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *PMDResult) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d", this.Data.Errors)
	return &tool.Summary{this.GetName(), body}
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

func NewResult(fileId bson.ObjectId, data []byte) (res *PMDResult, err error) {
	res = &PMDResult{Id: bson.NewObjectId(), FileId: fileId, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *PMDReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		return
	}
	res.Id = id
	res.Errors = 0
	for _, f := range res.Files{
		res.Errors += len(f.Violations)
	}
	return
}

type PMDReport struct {
	Id      bson.ObjectId
	Version string  `xml:"version,attr"`
	Files   []*File `xml:"file"`
	Errors int
}
type File struct {
	Name       string       `xml:"name,attr"`
	Violations []*Violation `xml:"violation"`
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
