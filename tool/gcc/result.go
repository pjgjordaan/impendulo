package gcc

import (
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type (
	Result struct {
		Id     bson.ObjectId `bson:"_id"`
		FileId bson.ObjectId `bson:"fileid"`
		Name   string        `bson:"name"`
		Report *Report       `bson:"report"`
		GridFS bool          `bson:"gridfs"`
		Type   string        `bson:"type"`
	}
)

func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Result) Summary() *tool.Summary {
	return &tool.Summary{}

}

func (this *Result) GetName() string {
	return this.Name
}

func (this *Result) OnGridFS() bool {
	return this.GridFS
}

func (this *Result) GetReport() tool.Report {
	return this.Report
}

func (this *Result) SetReport(report tool.Report) {
	this.Report = report.(*Report)
}

//ChartVals
func (this *Result) ChartVals() []*tool.ChartVal {
	return []*tool.ChartVal{
		&tool.ChartVal{Name: "Errors", Y: float64(this.Report.Errors), FileId: this.FileId},
		&tool.ChartVal{Name: "Warnings", Y: float64(this.Report.Warnings), FileId: this.FileId},
	}
}

func (this *Result) Template() string {
	return "gccresult"
}

func (this *Result) GetType() string {
	return this.Type
}

func NewResult(fileId bson.ObjectId, data []byte) (ret tool.ToolResult, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	id := bson.NewObjectId()
	report, err := NewReport(id, data)
	if err != nil {
		return
	}
	ret = &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		Report: report,
		GridFS: gridFS,
		Type:   NAME,
	}
	return
}
