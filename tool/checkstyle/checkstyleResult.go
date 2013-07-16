package checkstyle

import (
	"encoding/xml"
	"github.com/godfried/impendulo/util"
	"html/template"
	"labix.org/v2/mgo/bson"
	"reflect"
)

const NAME = "Checkstyle"

type CheckstyleResult struct {
	Id     bson.ObjectId     "_id"
	FileId bson.ObjectId     "fileid"
	Time   int64             "time"
	Data   *CheckstyleReport "data"
}

func (this *CheckstyleResult) Equals(that *CheckstyleResult) bool {
	return reflect.DeepEqual(this, that)
}

func (this *CheckstyleResult) Name() string {
	return NAME
}

func (this *CheckstyleResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *CheckstyleResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *CheckstyleResult) String() string {
	return "Type: tool.java.CheckstyleResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *CheckstyleResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "checkstyleCurrent.html", this.Data
	} else {
		return "checkstyleNext.html", this.Data
	}
}

func (this *CheckstyleResult) Success() bool {
	return true
}

func NewResult(fileId bson.ObjectId, data []byte) (res *CheckstyleResult, err error) {
	res = &CheckstyleResult{Id: bson.NewObjectId(), FileId: fileId, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *CheckstyleReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		return
	}
	res.Id = id
	return
}

type CheckstyleReport struct {
	Id      bson.ObjectId
	Version string  `xml:"version,attr"`
	Files   []*File `xml:"file"`
}
type File struct {
	Name   string   `xml:"name,attr"`
	Errors []*Error `xml:"error"`
}

type Error struct {
	Line     int           `xml:"line,attr"`
	Column   int           `xml:"column,attr"`
	Severity string        `xml:"severity,attr"`
	Message  template.HTML `xml:"message,attr"`
	Source   string        `xml:"source,attr"`
}
