package tool

import(
	"labix.org/v2/mgo/bson"
"reflect"
"time"
)

//Result describes a tool or test's results for a given file.
type Result struct {
	Id      bson.ObjectId "_id"
	FileId  bson.ObjectId "fileid"
	ToolId  bson.ObjectId "toolId"
	Name    string        "name"
	OutName string        "outname"
	ErrName string        "errname"
	OutData []byte        "outdata"
	ErrData []byte        "errdata"
	Error   error         "error"
	Time    int64         "time"
}

func (this *Result) Equals(that *Result) bool {
	return reflect.DeepEqual(this, that)
}

//NewResult
func ToolResult(fileId bson.ObjectId, tool *Tool, outdata, errdata []byte, err error) *Result {
	return &Result{bson.NewObjectId(), fileId, tool.Id, tool.Name, tool.OutName, tool.ErrName, outdata, errdata, err, time.Now().UnixNano()}
}


//NewResult
func NewResult(fileId, toolId bson.ObjectId, name, outname, errname string, outdata, errdata []byte, err error) *Result {
	return &Result{bson.NewObjectId(), fileId, toolId, name, outname, errname, outdata, errdata, err, time.Now().UnixNano()}
}