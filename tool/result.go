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
	Name    string        "name"
	OutData []byte        "outdata"
	ErrData []byte        "errdata"
	Error   error         "error"
	Time    int64         "time"
}

func (this *Result) Equals(that *Result) bool {
	return reflect.DeepEqual(this, that)
}

func (this *Result) String()string{
	errString := "None"
	if this.Error != nil{
		errString = this.Error.Error()
	}
	return "File: "+this.FileId.String()+"; Name: "+this.Name+"; OutData:"+string(this.OutData)+"; ErrData:"+string(this.ErrData)+"; Error: "+errString
}

func (this *Result) Output()string{
	return string(this.OutData)
}

func (this *Result) Errors()string{
	return string(this.ErrData)
}

//NewResult
func NewResult(fileId bson.ObjectId, tool Tool, outdata, errdata []byte, err error) *Result {
	return &Result{bson.NewObjectId(), fileId, tool.GetName(), outdata, errdata, err, time.Now().UnixNano()}
}
