package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
)

const (
	NAME = "Findbugs"
)

type (
	//Result is a tool.ToolResult and a tool.DisplayResult.
	//It is used to store the output of running Findbugs.
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

//String
func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Data: %s",
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
	body := fmt.Sprintf("Bugs: %d", this.Data.Summary.BugCount)
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
		graphData[0] = tool.CreateChart("Findbugs All")
		graphData[1] = tool.CreateChart("Findbugs Priority 1")
		graphData[2] = tool.CreateChart("Findbugs Priority 2")
		graphData[3] = tool.CreateChart("Findbugs Priority 3")
	}
	yB := float64(this.Data.Summary.BugCount)
	y1 := float64(this.Data.Summary.Priority1)
	y2 := float64(this.Data.Summary.Priority2)
	y3 := float64(this.Data.Summary.Priority3)
	tool.AddCoords(graphData[0], x, yB)
	tool.AddCoords(graphData[1], x, y1)
	tool.AddCoords(graphData[2], x, y2)
	tool.AddCoords(graphData[3], x, y3)
	return math.Max(max, math.Max(y1, math.Max(y2, math.Max(y3, yB))))
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
