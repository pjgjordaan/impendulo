package gcc

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	NAME     = "GCC"
	ERRORS   = "Errors"
	WARNINGS = "Warnings"
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

func (r *Result) GetTestId() bson.ObjectId {
	return ""
}

func (r *Result) GetId() bson.ObjectId {
	return r.Id
}

func (r *Result) GetFileId() bson.ObjectId {
	return r.FileId
}

func (r *Result) GetName() string {
	return r.Name
}

func (r *Result) OnGridFS() bool {
	return r.GridFS
}

func (r *Result) Reporter() result.Reporter {
	return r.Report
}

func (r *Result) SetReport(report result.Reporter) {
	r.Report = report.(*Report)
}

//Values
func (r *Result) Values() []*result.Value {
	return []*result.Value{
		&result.Value{Name: ERRORS, V: float64(r.Report.Errors), FileId: r.FileId},
		&result.Value{Name: WARNINGS, V: float64(r.Report.Warnings), FileId: r.FileId},
	}
}

func (r *Result) Value(n string) (*result.Value, error) {
	switch n {
	case ERRORS:
		return &result.Value{Name: ERRORS, V: float64(r.Report.Errors), FileId: r.FileId}, nil
	case WARNINGS:
		return &result.Value{Name: WARNINGS, V: float64(r.Report.Warnings), FileId: r.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown Value %s", n)
	}
}

func (r *Result) Types() []string {
	return []string{ERRORS, WARNINGS}
}

func (r *Result) Template() string {
	return "gccresult"
}

func (r *Result) GetType() string {
	return r.Type
}

func NewResult(fileId bson.ObjectId, data []byte) (result.Tooler, error) {
	id := bson.NewObjectId()
	report, e := NewReport(id, data)
	if e != nil {
		return nil, e
	}
	return &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		Report: report,
		GridFS: len(data) > tool.MAX_SIZE,
		Type:   NAME,
	}, nil
}
