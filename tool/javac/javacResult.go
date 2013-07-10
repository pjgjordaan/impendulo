package javac

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
)

const NAME = "Javac"

type JavacResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data   []byte        "data"
}

func (this *JavacResult) Equals(that *JavacResult) bool {
	return reflect.DeepEqual(this, that)
}

func (this *JavacResult) Name() string{
	return NAME
}

func (this *JavacResult) GetId() bson.ObjectId{
	return this.Id
}

func (this *JavacResult) GetFileId() bson.ObjectId{
	return this.FileId
}

func (this *JavacResult) String() string {
	return "Type: tool.java.JavacResult; Id: "+this.Id.Hex()+"; FileId: "+this.FileId.Hex() + "; Time: "+ util.Date(this.Time)
}

func (this *JavacResult) TemplateArgs(current bool)(string, interface{}){
	return "",""
}

func (this *JavacResult) Success() bool{
	return false
}

func NewResult(fileId bson.ObjectId, data []byte) *JavacResult{
	return &JavacResult{bson.NewObjectId(), fileId, util.CurMilis(), data}
}