package junit

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

const NAME = "JUnit Test"

type JUnitResult struct {
	Id       bson.ObjectId "_id"
	FileId   bson.ObjectId "fileid"
	TestName string        "name"
	Time     int64         "time"
	Data     []byte        "data"
}

func (this *JUnitResult) GetName() string {
	return this.TestName
}

func (this *JUnitResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *JUnitResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *JUnitResult) String() string {
	return "Type: tool.junit.JUnitResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *JUnitResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "junitCurrent", string(this.Data)
	} else {
		return "junitNext", string(this.Data)
	}
}

func (this *JUnitResult) Success() bool {
	return false
}

func NewResult(fileId bson.ObjectId, name string, data []byte) *JUnitResult {
	return &JUnitResult{bson.NewObjectId(), fileId, name, util.CurMilis(), data}
}
