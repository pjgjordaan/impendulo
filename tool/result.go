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