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
	TemplateArgs(current bool) (string, interface{})
	GetName() string
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

func (this *ErrorResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "errorCurrent", this.err
	} else {
		return "errorNext", this.err
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

func (this *CodeResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "codeCurrent", this.data
	} else {
		return "codeNext", this.data
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

func (this *SummaryResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "summaryCurrent", this.summary
	} else {
		return "summaryNext", this.summary
	}
}

func (this *SummaryResult) AddSummary(result ToolResult) {
	this.summary = append(this.summary, result.GetSummary())
}

type Summary struct{
	Name string
	Body string
}