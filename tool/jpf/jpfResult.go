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
	Threads []Thread `xml:"live-threads>thread"`
	Trace []Transition `xml:"trace>transition"`
	Result Result `xml:"result"`
	Stats Statistics `xml:"statistics"`
	Success bool
}

type Thread struct{
	Frames []string `xml:"frame"`
	RequestLocks []string  `xml:"lock-request object,attr"`
	OwnedLocks []string  `xml:"lock-owned object,attr"`
	Status string `xml:"status,attr"`
	Id int `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type Transition struct{
	Id int `xml:"id,attr"`
	ThreadId int `xml:"thread,attr"`
	CG ChoiceGenerator `xml:"cg"`
	Insns []Instruction `xml:"insn"`
}

func (this Transition) String()string{
	return `Transition Id: `+strconv.Itoa(this.Id)+`; Thread Id: `+strconv.Itoa(this.ThreadId)
}

type ChoiceGenerator struct{
	Class string `xml:"class,attr"`
	Choice string `xml:"choice,attr"`
}

type Instruction struct{
	Source string `xml:"src,attr"`
	Value string `xml:",innerxml"`
} 

type Result struct{
	Findings string `xml:"findings,attr"`
	Errors []Error `xml:"error"`
}

type Error struct{
	Id int `xml:"id,attr"`
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
	if res.Result.Findings == "none"{
		res.Success = true
	} else{
		res.Success = false
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
	buffer.WriteString(`<div class="accordion" id=jpfaccordion>`)
	buffer.WriteString(`<div class="accordion-group">`)
	buffer.WriteString(`<div class="accordion-heading">`)
	buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#jpfaccordion" href="#stats">Statistics.</a></div>`)
	buffer.WriteString(`<div id="stats" class="accordion-body collapse">`)
	buffer.WriteString(`<div class="accordion-inner">`)
	buffer.WriteString(`<dl class="dl-horizontal">`)
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
	buffer.WriteString(`</dl></div></div></div>`)	
	buffer.WriteString(`<div class="accordion-group">`)
	buffer.WriteString(`<div class="accordion-heading">`)
	buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#jpfaccordion" href="#errors">Errors.</a></div>`)
	buffer.WriteString(`<div id="errors" class="accordion-body collapse">`)
	buffer.WriteString(`<div class="accordion-inner">`)
	if this.Report.Success{
		buffer.WriteString(`<h3 class="text-success">No errors detected</h3>`)
	} else{
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
	}
	buffer.WriteString(`</div></div></div>`)
	buffer.WriteString(`<div class="accordion-group">`)
	buffer.WriteString(`<div class="accordion-heading">`)
	buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#jpfaccordion" href="#threads">Threads.</a></div>`)
	buffer.WriteString(`<div id="threads" class="accordion-body collapse">`)
	buffer.WriteString(`<div class="accordion-inner">`)
	for _, th := range this.Report.Threads{
		buffer.WriteString(`<dl class="dl-horizontal">`)
		buffer.WriteString(`<dt>Name</dt>`)
		buffer.WriteString(`<dd>`+th.Name+`</dd>`)
		buffer.WriteString(`<dt>Identifier</dt>`)
		buffer.WriteString(`<dd>`+strconv.Itoa(th.Id)+`</dd>`)
		buffer.WriteString(`<dt>Status</dt>`)
		buffer.WriteString(`<dd>`+th.Status+`</dd>`)
		buffer.WriteString(`<dt>Stack Frames</dt>`)
		for _, frame := range th.Frames{
			buffer.WriteString(`<dd>`+frame+`</dd>`)
		}
		buffer.WriteString(`</dl>`)
	}
	buffer.WriteString(`</div></div></div>`)
	buffer.WriteString(`<div class="accordion-group">`)
	buffer.WriteString(`<div class="accordion-heading">`)
	buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#jpfaccordion" href="#trace">Trace.</a></div>`)
	buffer.WriteString(`<div id="trace" class="accordion-body collapse">`)
	buffer.WriteString(`<div class="accordion-inner">`)
	buffer.WriteString(`<div class="accordion" id=traceaccordion>`)
	for _, trans := range this.Report.Trace{
		buffer.WriteString(`<div class="accordion-group">`)
		buffer.WriteString(`<div class="accordion-heading">`)
		buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#traceaccordion" href="#trans`+strconv.Itoa(trans.Id)+`">`)
		buffer.WriteString(trans.String())
		buffer.WriteString(`</a></div>`)
		buffer.WriteString(`<div id="trans`+strconv.Itoa(trans.Id)+`" class="accordion-body collapse">`)
		buffer.WriteString(`<div class="accordion-inner">`)
		buffer.WriteString(`<dl class="dl-horizontal">`)
		buffer.WriteString(`<dt>Choice Generator</dt>`)
		buffer.WriteString(`<dd>`+trans.CG.Class+`; choice = `+trans.CG.Choice+`</dd>`)
		buffer.WriteString(`<dt>Instructions</dt>`)
		for _, insn := range trans.Insns{
			buffer.WriteString(`<dd>`+insn.Value+` (`+insn.Source+`)</dd>`)
		}
		buffer.WriteString(`</dl>`)
		buffer.WriteString(`</div>`)
		buffer.WriteString(`</div>`)
		buffer.WriteString(`</div>`)
	}
	buffer.WriteString(`</div>`)
	buffer.WriteString(`</div></div></div>`)
	return template.HTML(buffer.String())
}

func (this *JPFResult) Success() bool{
	return false
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *JPFResult {
	return &JPFResult{bson.NewObjectId(), fileId, util.CurMilis(), ReadReport(data)}
}