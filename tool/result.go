package tool

import (
	"labix.org/v2/mgo/bson"
	"strings"
)

//Result describes a tool or test's results for a given file.
type Result interface {
	TemplateArgs(current bool) (string, interface{})
	Success() bool
	GetName() string
	GetId() bson.ObjectId
	GetFileId() bson.ObjectId
	String() string
}

func NewErrorResult(err error) *ErrorResult {
	return &ErrorResult{err}
}

type ErrorResult struct {
	err error
}

func (this *ErrorResult) String() string {
	return this.err.Error()
}

func (this *ErrorResult) GetName() string {
	return "Error"
}

func (this *ErrorResult) Success() bool {
	return false
}

func (this *ErrorResult) GetId() bson.ObjectId {
	return bson.NewObjectId()
}

func (this *ErrorResult) GetFileId() bson.ObjectId {
	return bson.NewObjectId()
}

func (this *ErrorResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "errorCurrent", this.err
	} else {
		return "errorNext", this.err
	}
}

func NewCodeResult(fileId bson.ObjectId, data []byte) *CodeResult {
	return &CodeResult{fileId, strings.TrimSpace(string(data))}
}

type CodeResult struct {
	fileId bson.ObjectId
	data   string
}

func (this *CodeResult) String() string {
	return this.GetName()
}

func (this *CodeResult) GetName() string {
	return CODE
}

func (this *CodeResult) Success() bool {
	return true
}

func (this *CodeResult) GetId() bson.ObjectId {
	return bson.NewObjectId()
}

func (this *CodeResult) GetFileId() bson.ObjectId {
	return this.fileId
}

func (this *CodeResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "codeCurrent", this.data
	} else {
		return "codeNext", this.data
	}
}

const CODE = "Code"
