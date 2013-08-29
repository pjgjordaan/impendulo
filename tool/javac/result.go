package javac

import (
	"bytes"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

const NAME = "Javac"

type Result struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	Data   []byte        "data"
}

func (this *Result) GetName() string {
	return this.Name
}

func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Result) GetSummary() *tool.Summary {
	var body string
	if this.Success() {
		body = "Compiled successfully."
	} else {
		body = "No compile."
	}
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *Result) GetData() interface{} {
	return this
}

func (this *Result) Template(current bool) string {
	if current {
		return "javacCurrent"
	} else {
		return "javacNext"
	}
}

func (this *Result) Success() bool {
	return bytes.Equal(this.Data, []byte("Compiled successfully"))
}

func (this *Result) Result() string {
	return string(this.Data)
}

func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("Compilation")
	}
	var y float64
	if this.Success() {
		y = 1
	} else {
		y = 0
	}
	tool.AddCoords(graphData[0], x, y)
	return 1
}

func NewResult(fileId bson.ObjectId, data []byte) *Result {
	return &Result{
		Id: bson.NewObjectId(), 
		FileId: fileId, 
		Name: NAME, 
		Data: data,
	}
}
