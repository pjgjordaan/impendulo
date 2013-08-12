package jpf

import (
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"strconv"
)

const NAME = "JPF"

type JPFResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Time   int64         "time"
	Data   *JPFReport    "data"
}

func (this *JPFResult) GetName() string {
	return NAME
}

func (this *JPFResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *JPFResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *JPFResult) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Result: %s \n Errors: %d", this.Data.Result.Findings, len(this.Data.Result.Errors))
	return &tool.Summary{this.GetName(), body}
}

func (this *JPFResult) Success() bool {
	return true
}

func (this *JPFResult) GetData() interface{} {
	return this.Data
}

func (this *JPFResult) Template(current bool) string {
	if current {
		return "jpfCurrent"
	} else {
		return "jpfNext"
	}
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) (res *JPFResult, err error) {
	res = &JPFResult{Id: bson.NewObjectId(), FileId: fileId, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

type JPFReport struct {
	Id      bson.ObjectId
	Version string       `xml:"jpf-version"`
	Threads []Thread     `xml:"live-threads>thread"`
	Trace   []Transition `xml:"trace>transition"`
	Result  Result       `xml:"result"`
	Stats   Statistics   `xml:"statistics"`
	Success bool
}

type Thread struct {
	Frames       []string `xml:"frame"`
	RequestLocks []string `xml:"lock-request object,attr"`
	OwnedLocks   []string `xml:"lock-owned object,attr"`
	Status       string   `xml:"status,attr"`
	Id           int      `xml:"id,attr"`
	Name         string   `xml:"name,attr"`
}

type Transition struct {
	Id       int             `xml:"id,attr"`
	ThreadId int             `xml:"thread,attr"`
	CG       ChoiceGenerator `xml:"cg"`
	Insns    []Instruction   `xml:"insn"`
}

func (this Transition) String() string {
	return `Transition Id: ` + strconv.Itoa(this.Id) + `; Thread Id: ` + strconv.Itoa(this.ThreadId)
}

type ChoiceGenerator struct {
	Class  string `xml:"class,attr"`
	Choice string `xml:"choice,attr"`
}

type Instruction struct {
	Source string `xml:"src,attr"`
	Value  string `xml:",innerxml"`
}

type Result struct {
	Findings string  `xml:"findings,attr"`
	Errors   []Error `xml:"error"`
}

type Error struct {
	Id       int    `xml:"id,attr"`
	Property string `xml:"property"`
	Details  string `xml:"details"`
}

type Statistics struct {
	Time              string `xml:"elapsed-time"`
	NewStates         int    `xml:"new-states"`
	VisitedStates     int    `xml:"visited-states"`
	BacktrackedStates int    `xml:"backtracked-states"`
	EndStates         int    `xml:"end-states"`
	Memory            int    `xml:"max-memory"`
}

func genReport(id bson.ObjectId, data []byte) (res *JPFReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		return
	}
	if res.Result.Findings == "none" {
		res.Success = true
	} else {
		res.Success = false
	}
	res.Id = id
	return
}
