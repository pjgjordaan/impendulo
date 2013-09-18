package junit

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "JUnit Test"
)

type (
	//Result is an implementation of ToolResult, DisplayResult
	//and GraphResult for JUnit test results.
	Result struct {
		Id       bson.ObjectId "_id"
		FileId   bson.ObjectId "fileid"
		TestName string        "name"
		Data     *Report       "data"
		GridFS   bool          "gridfs"
	}
)

func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

func (this *Result) OnGridFS() bool {
	return this.GridFS
}

func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Data: %s",
		this.Id, this.FileId, this.TestName, this.Data)
}

func (this *Result) GetName() string {
	return this.TestName
}

func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Result) Summary() *tool.Summary {
	body := fmt.Sprintf("Tests: %d \n Failures: %d \n Errors: %d \n Time: %f",
		this.Data.Tests, this.Data.Failures, this.Data.Errors, this.Data.Time)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *Result) GetData() interface{} {
	return this.Data
}

func (this *Result) Success() bool {
	return this.Data.Success()
}

func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart(this.GetName() + " Errors")
		graphData[1] = tool.CreateChart(this.GetName() + " Failures")
		graphData[2] = tool.CreateChart(this.GetName() + " Successes")
	}
	yE := float64(this.Data.Errors)
	yF := float64(this.Data.Failures)
	yS := float64(this.Data.Tests - this.Data.Failures - this.Data.Errors)
	tool.AddCoords(graphData[0], x, yE)
	tool.AddCoords(graphData[1], x, yF)
	tool.AddCoords(graphData[2], x, yS)
	return math.Max(max, math.Max(yE, math.Max(yF, yS)))
}

func NewResult(fileId bson.ObjectId, name string, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:       bson.NewObjectId(),
		FileId:   fileId,
		TestName: name,
		GridFS:   gridFS,
	}
	res.Data, err = NewReport(res.Id, data)
	return
}
