package jpf

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "JPF"
)

type (
	//Result is a JPF implementation of tool.ToolResult, tool.DisplayResult and tool.GraphResult.
	Result struct {
		Id     bson.ObjectId "_id"
		FileId bson.ObjectId "fileid"
		Name   string        "name"
		Data   *Report       "data"
		GridFS bool          "gridfs"
	}
)

//SetData is used to change this result's data. This comes in handy
//when putting data into/getting data out of GridFS
func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

//OnGridFS determines whether this structure is partially stored on the
//GridFS.
func (this *Result) OnGridFS() bool {
	return this.GridFS
}

//String
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nData: %s\n",
		this.Id, this.FileId, this.Name, this.Data)
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
	body := fmt.Sprintf("Result: %s \n Errors: %d",
		this.Data.Findings.Description, len(this.Data.Findings.Errors))
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

//Success
func (this *Result) Success() bool {
	return true
}

//GetData
func (this *Result) GetData() interface{} {
	return this.Data
}

//AddGraphData
func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("JPF Error Detection Time")
	}
	var yT float64
	if this.Data.Errors() == 0 {
		yT = 0
	} else {
		yT = float64(this.Data.Stats.Time)
	}
	tool.AddCoords(graphData[0], x, yT)
	return math.Max(max, yT)
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
	}
	res.Data, err = NewReport(res.Id, data)
	return
}
