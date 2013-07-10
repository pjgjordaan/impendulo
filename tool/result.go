package tool

import (
	"labix.org/v2/mgo/bson"
	"strings"
)

//Result describes a tool or test's results for a given file.
type Result interface {
	TemplateArgs(current bool) (string, interface{})
	Success() bool
	Name() string
	GetId() bson.ObjectId
	GetFileId() bson.ObjectId
	String() string
}

func NewErrorResult(msg string)*ErrorResult{
	return &ErrorResult{msg}
}

type ErrorResult struct{
	msg string
}

func (this *ErrorResult) String()string{
	return this.msg
}

func (this *ErrorResult) Name()string{
	return "Error"
}

func (this *ErrorResult) Success()bool{
	return false
}

func (this *ErrorResult) GetId()bson.ObjectId{
	return bson.NewObjectId()
}

func (this *ErrorResult) GetFileId()bson.ObjectId{
	return bson.NewObjectId()
}

func (this *ErrorResult) TemplateArgs(current bool)(string, interface{}){
	if current{
		return "errorCurrent.html", this.msg
	} else{
		return "errorNext.html", this.msg
	} 
}

func NewCodeResult(fileId bson.ObjectId, data []byte)*CodeResult{
	return &CodeResult{fileId, strings.TrimSpace(string(data))}
}

type CodeResult struct{
	fileId bson.ObjectId
	data string
}

func (this *CodeResult) String()string{
	return this.Name()
}

func (this *CodeResult) Name()string{
	return "Code"
}

func (this *CodeResult) Success()bool{
	return true
}

func (this *CodeResult) GetId()bson.ObjectId{
	return bson.NewObjectId()
}

func (this *CodeResult) GetFileId()bson.ObjectId{
	return this.fileId
}

func (this *CodeResult) TemplateArgs(current bool)(string, interface{}){
	if current{
		return "codeCurrent.html", this.data
	} else{
		return "codeNext.html", this.data
	} 
}