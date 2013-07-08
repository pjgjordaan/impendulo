package jpf

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
	"html/template"
	"encoding/xml"
	"bytes"
	"strconv"
	"strings"
)

const NAME = "JPF"

type JPFReport struct{
	Version string `xml:"jpf-version"`
	Threads []Thread `xml:"live-threads>thread`
	Result Result `xml:"result"`
	Stats Statistics `xml:"statistics"`
}

type Thread struct{
	Frames []string `xml:"frame"`
	Status string `xml:"status,attr"`
	Id string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type Result struct{
	Findings string `xml:"findings,attr"`
	Errors []Error `xml:"error"`
}

type Error struct{
	Id string `xml:"id,attr"`
	Property string `xml:"property"`
	Details string `xml:"details"`
}

type Statistics struct{
	Time string `xml:"elapsed-time"`
	NewStates int `xml:"new-states"`
	VisitedStates int `xml:"visited-states"`
	BacktrackedStates int `xml:"backtracked-states"`
	EndStates int `xml:"end-states"`
	Memory int `xml:"max-memory"`
}

func ReadReport(data []byte)(res *JPFReport) {
	if err := xml.Unmarshal(data, &res); err != nil{
		panic(err)
	}
	return
}

type JPFResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Report *JPFReport    "report"
}

func (this *JPFResult) Equals(that *JPFResult) bool {
	return reflect.DeepEqual(this, that)
}

func (this *JPFResult) Name() string{
	return NAME
}

func (this *JPFResult) GetId() bson.ObjectId{
	return this.Id
}

func (this *JPFResult) GetFileId() bson.ObjectId{
	return this.FileId
}

func (this *JPFResult) String() string {
	return "Type: tool.jpf.JPFResult; Id: "+this.Id.Hex()+"; FileId: "+this.FileId.Hex() + "; Time: "+ util.Date(this.Time)
}

func (this *JPFResult) HTML()template.HTML{
	buffer := new(bytes.Buffer)
	buffer.WriteString(`<h3 class="text-error">`)
	buffer.WriteString(strconv.Itoa(len(this.Report.Result.Errors))+" errors found.")
	buffer.WriteString("</h3>")
	for _, err := range this.Report.Result.Errors{
		buffer.WriteString(`<h4 class="text-warning">`)
		buffer.WriteString(err.Property)
		buffer.WriteString("</h4>")
		buffer.WriteString("<p>")
		details := strings.Replace(err.Details, "\n", "<br>", -1)
		buffer.WriteString(details)
		buffer.WriteString("</p>")
	}
	buffer.WriteString(`<h3 class="text-info">Statistics.</h3>`)
	buffer.WriteString(`<dl class="text-info dl-horizontal">`)
	buffer.WriteString(`<dt>Elapsed time</dt>`)
	buffer.WriteString(`<dd>`+this.Report.Stats.Time+`</dd>`)
	buffer.WriteString(`<dt>New States</dt>`)
	buffer.WriteString(`<dd>`+strconv.Itoa(this.Report.Stats.NewStates)+`</dd>`)
	buffer.WriteString(`<dt>Visited States</dt>`)
	buffer.WriteString(`<dd>`+strconv.Itoa(this.Report.Stats.VisitedStates)+`</dd>`)
	buffer.WriteString(`<dt>BackTracked States</dt>`)
	buffer.WriteString(`<dd>`+strconv.Itoa(this.Report.Stats.BacktrackedStates)+`</dd>`)
	buffer.WriteString(`<dt>End States</dt>`)
	buffer.WriteString(`<dd>`+strconv.Itoa(this.Report.Stats.EndStates)+`</dd>`)
	buffer.WriteString(`<dt>Memory Usage</dt>`)
	buffer.WriteString(`<dd>`+strconv.Itoa(this.Report.Stats.Memory)+` MB</dd>`)
	buffer.WriteString(`</dl>`)	
	return template.HTML(buffer.String())
}

func (this *JPFResult) Success() bool{
	return false
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *JPFResult {
	return &JPFResult{bson.NewObjectId(), fileId, util.CurMilis(), ReadReport(data)}
}