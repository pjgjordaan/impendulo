package jpf

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
	"encoding/xml"
	"strconv"
)

const NAME = "JPF"

type JPFResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data *JPFReport "data"
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

func (this *JPFResult) Success() bool{
	return false
}

func (this *JPFResult) TemplateArgs(current bool) (string, interface{}){
	if current{
		return "jpfCurrent.html", this.Data
	}else{
		return "jpfNext.html", this.Data
	}
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) *JPFResult {
	id := bson.NewObjectId()
	return &JPFResult{id, fileId, util.CurMilis(), genReport(id, data)}
}


type JPFReport struct{
	Id bson.ObjectId
	Version string `xml:"jpf-version"`
	Threads []Thread `xml:"live-threads>thread"`
	Trace []Transition `xml:"trace>transition"`
	Result Result `xml:"result"`
	Stats Statistics `xml:"statistics"`
	Success bool
}

type Thread struct{
	Frames []string `xml:"frame"`
	RequestLocks []string  `xml:"lock-request object,attr"`
	OwnedLocks []string  `xml:"lock-owned object,attr"`
	Status string `xml:"status,attr"`
	Id int `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type Transition struct{
	Id int `xml:"id,attr"`
	ThreadId int `xml:"thread,attr"`
	CG ChoiceGenerator `xml:"cg"`
	Insns []Instruction `xml:"insn"`
}

func (this Transition) String()string{
	return `Transition Id: `+strconv.Itoa(this.Id)+`; Thread Id: `+strconv.Itoa(this.ThreadId)
}

type ChoiceGenerator struct{
	Class string `xml:"class,attr"`
	Choice string `xml:"choice,attr"`
}

type Instruction struct{
	Source string `xml:"src,attr"`
	Value string `xml:",innerxml"`
} 

type Result struct{
	Findings string `xml:"findings,attr"`
	Errors []Error `xml:"error"`
}

type Error struct{
	Id int `xml:"id,attr"`
	Property string `xml:"property"`
	Details string `xml:"details"`
}

type Statistics struct{
	Time string `xml:"elapsed-time"`
	NewStates int `xml:"new-states"`
	VisitedStates int `xml:"visited-states"`
	BacktrackedStates int `xml:"backtracked-states"`
	EndStates int `xml:"end-states"`
	Memory int `xml:"max-memory"`
}

func genReport(id bson.ObjectId, data []byte)(res *JPFReport) {
	if err := xml.Unmarshal(data, &res); err != nil{
		panic(err)
	}
	if res.Result.Findings == "none"{
		res.Success = true
	} else{
		res.Success = false
	}
	res.Id = id
	return
}
