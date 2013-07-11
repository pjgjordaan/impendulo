package lint4j

import (
	"bytes"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"reflect"
)

const NAME = "Lint4j"

type Lint4jResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data   []byte        "data"
}

func (this *Lint4jResult) Equals(that *Lint4jResult) bool {
	return reflect.DeepEqual(this, that)
}

func (this *Lint4jResult) Name() string {
	return NAME
}

func (this *Lint4jResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *Lint4jResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Lint4jResult) String() string {
	return "Type: tool.java.Lint4jResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *Lint4jResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "findbugsCurrent.html", this
	} else {
		return "findbugsNext.html", this
	}
}

func (this *Lint4jResult) Success() bool {
	return bytes.Equal(this.Data, []byte("Compiled successfully"))
}

func (this *Lint4jResult) Result() string {
	return string(this.Data)
}

func NewResult(fileId bson.ObjectId, data []byte) tool.Result {
	return &Lint4jResult{bson.NewObjectId(), fileId, util.CurMilis(), data}
}
