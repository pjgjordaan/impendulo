package javac

import (
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "Javac"
)

type (
	//Result is a Javac implementation of tool.ToolResult, tool.DisplayResult
	//and tool.GraphResult. It contains the result of compiling a Java source
	//file using the specified Java compiler.
	Result struct {
		Id     bson.ObjectId "_id"
		FileId bson.ObjectId "fileid"
		Name   string        "name"
		Data   *Report       "data"
		GridFS bool          "gridfs"
	}
)

//SetData
func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

//OnGridFS
func (this *Result) OnGridFS() bool {
	return this.GridFS
}

//GetName
func (this *Result) GetName() string {
	return this.Name
}

//GetId
func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

//GetFileId
func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

//Summary
func (this *Result) Summary() *tool.Summary {
	var body string
	if this.Data.Success() {
		body = "Compiled successfully."
	} else {
		body = "No compile."
	}
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//GetData
func (this *Result) GetData() interface{} {
	return this.Data
}

//AddGraphData
func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart(this.GetName() + " Errors")
		graphData[1] = tool.CreateChart(this.GetName() + " Warnings")
	}
	yE, yW := 0.0, 0.0
	if this.Data.Errors() {
		yE = float64(this.Data.Count)
	} else if this.Data.Warnings() {
		yW = float64(this.Data.Count)
	}
	tool.AddCoords(graphData[0], x, yE)
	tool.AddCoords(graphData[1], x, yW)
	return math.Max(max, math.Max(yE, yW))
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *Result {
	gridFS := len(data) > tool.MAX_SIZE
	id := bson.NewObjectId()
	return &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
		Data:   NewReport(id, data),
	}
}
