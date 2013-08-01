package tool

import (
	"labix.org/v2/mgo/bson"
	"strings"
)

type ToolResult interface {
	GetId() bson.ObjectId
	GetFileId() bson.ObjectId
	GetSummary() *Summary
	GetName() string
}

//Result describes a tool or test's results for a given file.
type DisplayResult interface {
	Template(current bool) string
	GetName() string
	GetData() interface{}
}

func NewErrorResult(fileId bson.ObjectId, err error) *ErrorResult {
	return &ErrorResult{fileId, err}
}

type ErrorResult struct {
	fileId bson.ObjectId
	err error
}

func (this *ErrorResult) GetName() string {
	return "Error"
}

func (this *ErrorResult) GetData() interface{} {
	return this.err
}

func (this *ErrorResult) Template(current bool) string {
	if current {
		return "errorCurrent"
	} else {
		return "errorNext"
	}
}

const CODE = "Code"

func NewCodeResult(fileId bson.ObjectId, data []byte) *CodeResult {
	return &CodeResult{fileId, strings.TrimSpace(string(data))}
}

type CodeResult struct {
	fileId bson.ObjectId
	data   string
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

const SUMMARY = "Summary"

func NewSummaryResult() *SummaryResult {
	return &SummaryResult{make([]*Summary,0)}
}

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

type Summary struct{
	Name string
	Body string
}