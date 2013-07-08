package jpf

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
	"html/template"
	"encoding/json"
)

const NAME = "JPF"

type JPFError struct{
	Details string "details"
	Description string "description"
}

func ReadErrors(data []byte)( res []*JPFError) {
	if err := json.Unmarshal(data, &res); err != nil{
		panic(err)
	}
	return
}

type JPFResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Errors   []*JPFError        "errors"
}

func (this *JPFResult) Equals(that *JPFResult) bool {
	return reflect.DeepEqual(this, that)
}

func (this *JPFResult) Name() string{
	return NAME
}

func (this *JPFResult) GetId() bson.ObjectId{
	return this.Id
}

func (this *JPFResult) GetFileId() bson.ObjectId{
	return this.FileId
}

func (this *JPFResult) String() string {
	return "Type: tool.jpf.JPFResult; Id: "+this.Id.Hex()+"; FileId: "+this.FileId.Hex() + "; Time: "+ util.Date(this.Time)
}

func (this *JPFResult) HTML()template.HTML{
	return template.HTML(this.Name())
}

func (this *JPFResult) Success() bool{
	return false
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *JPFResult {
	return &JPFResult{bson.NewObjectId(), fileId, util.CurMilis(), ReadErrors(data)}
}