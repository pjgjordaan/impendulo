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

type JPFResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Report *JPFReport    "report"
	HTML template.HTML "html"
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


func (this *JPFResult) GetHTML() template.HTML{
	return this.HTML
}


func (this *JPFResult) String() string {
	return "Type: tool.jpf.JPFResult; Id: "+this.Id.Hex()+"; FileId: "+this.FileId.Hex() + "; Time: "+ util.Date(this.Time)
}


func (this *JPFResult) Success() bool{
	return false
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *JPFResult {
	report := ReadReport(data)
	gen := &Generator{id: fileId.Hex()}
	html := gen.genHTML(report)
	return &JPFResult{bson.NewObjectId(), fileId, util.CurMilis(), report, html}
}


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

type Generator struct{
	id string
	buffer *bytes.Buffer
}

func (this *Generator) genHTML(report *JPFReport)template.HTML{
	this.buffer = new(bytes.Buffer)
	this.buffer.WriteString(`<div class="accordion" id=jpfaccordion`+this.id+`>`)
	this.genStats(report.Stats)
	this.genErrors(report.Result.Errors, report.Success)
	this.genThreads(report.Threads)
	this.genTrace(report.Trace)
	this.buffer.WriteString(`</div>`)
	return template.HTML(this.buffer.String()) 
}

func (this *Generator) genHeader(name string){
	this.buffer.WriteString(`<div class="accordion-group">`)
	this.buffer.WriteString(`<div class="accordion-heading">`)
	this.buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#jpfaccordion`+this.id+`" href="#`+name+this.id+`">`+name+`</a></div>`)
	this.buffer.WriteString(`<div id="`+name+this.id+`" class="accordion-body collapse">`)
	this.buffer.WriteString(`<div class="accordion-inner">`)
}

func (this *Generator) genFooter(){
	this.buffer.WriteString(`</div></div></div>`)
}

func (this *Generator) genStats(stats Statistics){
	this.genHeader("Statistics")
	this.buffer.WriteString(`<dl class="dl-horizontal">`)
	this.buffer.WriteString(`<dt>Elapsed time</dt>`)
	this.buffer.WriteString(`<dd>`+stats.Time+`</dd>`)
	this.buffer.WriteString(`<dt>New States</dt>`)
	this.buffer.WriteString(`<dd>`+strconv.Itoa(stats.NewStates)+`</dd>`)
	this.buffer.WriteString(`<dt>Visited States</dt>`)
	this.buffer.WriteString(`<dd>`+strconv.Itoa(stats.VisitedStates)+`</dd>`)
	this.buffer.WriteString(`<dt>BackTracked States</dt>`)
	this.buffer.WriteString(`<dd>`+strconv.Itoa(stats.BacktrackedStates)+`</dd>`)
	this.buffer.WriteString(`<dt>End States</dt>`)
	this.buffer.WriteString(`<dd>`+strconv.Itoa(stats.EndStates)+`</dd>`)
	this.buffer.WriteString(`<dt>Memory Usage</dt>`)
	this.buffer.WriteString(`<dd>`+strconv.Itoa(stats.Memory)+` MB</dd>`)
	this.buffer.WriteString(`</dl>`)
	this.genFooter()	
}

func (this *Generator) genErrors(errors []Error, success bool){
	this.genHeader("Errors")
	if success{
		this.buffer.WriteString(`<h3 class="text-success">No errors detected</h3>`)
	} else{
		this.buffer.WriteString(`<h3 class="text-error">`)
		this.buffer.WriteString(strconv.Itoa(len(errors))+" errors found.")
		this.buffer.WriteString("</h3>")
		for _, err := range errors{
			this.buffer.WriteString(`<h4 class="text-warning">`)
			this.buffer.WriteString(err.Property)
			this.buffer.WriteString("</h4>")
			this.buffer.WriteString("<p>")
			details := strings.Replace(err.Details, "\n", "<br>", -1)
			this.buffer.WriteString(details)
			this.buffer.WriteString("</p>")
		}
	}
	this.genFooter()
}

func (this *Generator) genThreads(threads []Thread){
	this.genHeader("Threads")
	for _, th := range threads{
		this.buffer.WriteString(`<dl class="dl-horizontal">`)
		this.buffer.WriteString(`<dt>Name</dt>`)
		this.buffer.WriteString(`<dd>`+th.Name+`</dd>`)
		this.buffer.WriteString(`<dt>Identifier</dt>`)
		this.buffer.WriteString(`<dd>`+strconv.Itoa(th.Id)+`</dd>`)
		this.buffer.WriteString(`<dt>Status</dt>`)
		this.buffer.WriteString(`<dd>`+th.Status+`</dd>`)
		this.buffer.WriteString(`<dt>Stack Frames</dt>`)
		for _, frame := range th.Frames{
			this.buffer.WriteString(`<dd>`+frame+`</dd>`)
		}
		this.buffer.WriteString(`</dl>`)
	}
	this.genFooter()
}

func (this *Generator) genTrace(trace []Transition){
	this.genHeader("Trace")
	this.buffer.WriteString(`<div class="accordion" id=traceaccordion`+this.id+`>`)
	for _, trans := range trace{
		this.buffer.WriteString(`<div class="accordion-group">`)
		this.buffer.WriteString(`<div class="accordion-heading">`)
		this.buffer.WriteString(`<a class="accordion-toggle" data-toggle="collapse" data-parent="#traceaccordion`+this.id+`" href="#`+this.id+`trans`+strconv.Itoa(trans.Id)+`">`)
		this.buffer.WriteString(trans.String())
		this.buffer.WriteString(`</a></div>`)
		this.buffer.WriteString(`<div id="`+this.id+`trans`+strconv.Itoa(trans.Id)+`" class="accordion-body collapse">`)
		this.buffer.WriteString(`<div class="accordion-inner">`)
		this.buffer.WriteString(`<dl class="dl-horizontal">`)
		this.buffer.WriteString(`<dt>Choice Generator</dt>`)
		this.buffer.WriteString(`<dd>`+trans.CG.Class+`; choice = `+trans.CG.Choice+`</dd>`)
		this.buffer.WriteString(`</dl>`)
		this.buffer.WriteString(`<h5>Instructions</h5>`)
		this.buffer.WriteString(`<ol>`)
		for _, insn := range trans.Insns{
			this.buffer.WriteString(`<dd>`+insn.Value+` (`+insn.Source+`)</dd>`)
		}
		this.buffer.WriteString(`</ol>`)
		this.genFooter()
	}
	this.buffer.WriteString(`</div>`)
	this.genFooter()
}