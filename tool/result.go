package tool

import (
	"labix.org/v2/mgo/bson"
	"strings"
	"fmt"
)

const (
	NORESULT = "No result"
	TIMEOUT  = "Timeout"
	SUMMARY  = "Summary"
	ERROR    = "Error"
	CODE     = "Code"
)


//GraphResult is used to display result data in a graph.
type GraphResult interface {
	AddGraphData(curMax float64, graphData []map[string]interface{}) (newMax float64)
	GetName() string
}

func AddCoords(chart map[string]interface{}, y float64){
	x := float64(len(chart["data"].([]map[string]float64)))
	chart["data"] = append(chart["data"].([]map[string]float64), 
		map[string]float64{"x": x, "y": y})
}

func CreateChart(name string)(chart map[string]interface{}){
	chart = make(map[string]interface{})
	chart["name"] = name
	chart["data"] = make([]map[string] float64, 0)
	return
}
		

//ToolResult is used to store tool result data.
type ToolResult interface {
	GetId() bson.ObjectId
	GetFileId() bson.ObjectId
	GetSummary() *Summary
	GetName() string
}

//DisplayResult is used to display result data.
type DisplayResult interface {
	Template(current bool) string
	GetName() string
	GetData() interface{}
}

//NoResult is a DisplayResult used to indicate that a
//Tool provided no result when run.
type NoResult struct{}

func (this *NoResult) GetName() string {
	return NORESULT
}

func (this *NoResult) GetData() interface{} {
	return "No output generated."
}

func (this *NoResult) Template(current bool) string {
	if current {
		return "errorCurrent"
	} else {
		return "errorNext"
	}
}

//TimeoutResult is a DisplayResult used to indicate that a
//Tool timed out when running.
type TimeoutResult struct{}

func (this *TimeoutResult) GetName() string {
	return TIMEOUT
}

func (this *TimeoutResult) GetData() interface{} {
	return "A timeout occured during tool execution."
}

func (this *TimeoutResult) Template(current bool) string {
	if current {
		return "errorCurrent"
	} else {
		return "errorNext"
	}
}

//ErrorResult is a DisplayResult used to indicate that an error
//occured when retrieving a Tool's result.
type ErrorResult struct {
	err error
}

func NewErrorResult(err error) *ErrorResult {
	return &ErrorResult{err}
}

func (this *ErrorResult) GetName() string {
	return ERROR
}

func (this *ErrorResult) GetData() interface{} {
	return this.err.Error()
}

func (this *ErrorResult) Template(current bool) string {
	if current {
		return "errorCurrent"
	} else {
		return "errorNext"
	}
}

func NewCodeResult(data []byte) *CodeResult {
	return &CodeResult{strings.TrimSpace(string(data))}
}

//CodeResult is a DisplayResult used to display a source file's code.
type CodeResult struct {
	data string
}

func (this *CodeResult) GetName() string {
	return CODE
}

func (this *CodeResult) GetData() interface{} {
	return this.data
}

func (this *CodeResult) Template(current bool) string {
	if current {
		return "codeCurrent"
	} else {
		return "codeNext"
	}
}

func NewSummaryResult() *SummaryResult {
	return &SummaryResult{make([]*Summary, 0)}
}

//SummaryResult is a DisplayResult used to
//provide a summary of all results.
type SummaryResult struct {
	summary []*Summary
}

func (this *SummaryResult) GetName() string {
	return SUMMARY
}

func (this *SummaryResult) GetData() interface{} {
	return this.summary
}

func (this *SummaryResult) Template(current bool) string {
	if current {
		return "summaryCurrent"
	} else {
		return "summaryNext"
	}
}

func (this *SummaryResult) AddSummary(result ToolResult) {
	this.summary = append(this.summary, result.GetSummary())
}

//Summary is short summary of a ToolResult's result.
type Summary struct {
	Name string
	Body string
}

type XMLError struct{
	err error
	origin string
}

func (this *XMLError) Error()string{
	return fmt.Sprintf("Encountered error %q while parsing xml in %s.", 
		this.err, this.origin)
}

func NewXMLError(err error, origin string) *XMLError{
	return &XMLError{err, origin}
}