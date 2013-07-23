package javac

import (
	"bytes"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

const NAME = "Javac"

type JavacResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data   []byte        "data"
}

func (this *JavacResult) GetName() string {
	return NAME
}

func (this *JavacResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *JavacResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *JavacResult) String() string {
	return "Type: tool.java.JavacResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *JavacResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "javacCurrent", this
	} else {
		return "javacNext", this
	}
}

func (this *JavacResult) Success() bool {
	return bytes.Equal(this.Data, []byte("Compiled successfully"))
}

func (this *JavacResult) Result() string {
	return string(this.Data)
}

func NewResult(fileId bson.ObjectId, data []byte) *JavacResult {
	return &JavacResult{bson.NewObjectId(), fileId, util.CurMilis(), data}
}
