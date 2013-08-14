package checkstyle

import (
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"html/template"
	"labix.org/v2/mgo/bson"
	"math"
)

const NAME = "Checkstyle"

type CheckstyleResult struct {
	Id     bson.ObjectId     "_id"
	FileId bson.ObjectId     "fileid"
	Name string "name"
	Time   int64             "time"
	Data   *CheckstyleReport "data"
}

func (this *CheckstyleResult) GetName() string {
	return this.Name
}

func (this *CheckstyleResult) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d",
		this.Data.Errors)
	return &tool.Summary{this.GetName(), body}
}

func (this *CheckstyleResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *CheckstyleResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *CheckstyleResult) GetData() interface{} {
	return this.Data
}

func (this *CheckstyleResult) Template(current bool) string {
	if current {
		return "checkstyleCurrent"
	} else {
		return "checkstyleNext"
	}
}

func (this *CheckstyleResult) Success() bool {
	return true
}

func (this *CheckstyleResult) AddGraphData(max float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil{
		graphData[0] = make(map[string]interface{})
		graphData[0]["name"] = "Checkstyle Errors"
		graphData[0]["data"] = make([]map[string] float64, 0)
	}
	x := float64(this.Time/1000)	
	y := float64(this.Data.Errors)
	graphData[0]["data"] = append(graphData[0]["data"].([]map[string]float64), map[string]float64{"x": x, "y": y})
	return math.Max(max, y)
}

func NewResult(fileId bson.ObjectId, data []byte) (res *CheckstyleResult, err error) {
	res = &CheckstyleResult{Id: bson.NewObjectId(), FileId: fileId, Name: NAME, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *CheckstyleReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		return
	}
	res.Id = id
	res.Errors = 0
	for _, f := range res.Files {
		res.Errors += len(f.Errors)
	}
	return
}

type CheckstyleReport struct {
	Id      bson.ObjectId
	Version string `xml:"version,attr"`
	Errors  int
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
