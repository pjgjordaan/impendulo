package pmd

import (
	"encoding/xml"
	"github.com/godfried/impendulo/util"
	"html/template"
	"labix.org/v2/mgo/bson"
	"reflect"
)

const NAME = "PMD"

type PMDResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data   *PMDReport    "data"
}

func (this *PMDResult) Equals(that *PMDResult) bool {
	return reflect.DeepEqual(this, that)
}

func (this *PMDResult) Name() string {
	return NAME
}

func (this *PMDResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *PMDResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *PMDResult) String() string {
	return "Type: tool.java.PMDResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *PMDResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "pmdCurrent.html", this.Data
	} else {
		return "pmdNext.html", this.Data
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
	return
}

type PMDReport struct {
	Id      bson.ObjectId
	Version string  `xml:"version,attr"`
	Files   []*File `xml:"file"`
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
