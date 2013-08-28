package jpf

import (
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
	"strconv"
)

const NAME = "JPF"

type JPFResult struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	Data   *JPFReport    "data"
}

func (this *JPFResult) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nData: %s\n",
		this.Id, this.FileId, this.Name, this.Data)
}

func (this *JPFResult) GetName() string {
	return this.Name
}

func (this *JPFResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *JPFResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *JPFResult) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Result: %s \n Errors: %d", this.Data.Result.Findings, len(this.Data.Result.Errors))
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
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

func (this *JPFResult) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("JPF New States")
		graphData[1] = tool.CreateChart("JPF Visited States")
		graphData[2] = tool.CreateChart("JPF Backtracked States")
		graphData[3] = tool.CreateChart("JPF End States")
	}
	yN := float64(this.Data.Stats.NewStates)
	yV := float64(this.Data.Stats.VisitedStates)
	yB := float64(this.Data.Stats.BacktrackedStates)
	yE := float64(this.Data.Stats.EndStates)
	tool.AddCoords(graphData[0], x, yN)
	tool.AddCoords(graphData[1], x, yV)
	tool.AddCoords(graphData[2], x, yB)
	tool.AddCoords(graphData[3], x, yE)
	return math.Max(max, math.Max(yN, math.Max(yV, math.Max(yB, yE))))
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) (res *JPFResult, err error) {
	res = &JPFResult{
		Id: bson.NewObjectId(), 
		FileId: fileId, 
		Name: NAME,
	}
	res.Data, err = genReport(res.Id, data)
	return
}

type JPFReport struct {
	Id      bson.ObjectId
	Version string       `xml:"jpf-version"`
	Threads []*Thread     `xml:"live-threads>thread"`
	Trace   []*Transition `xml:"trace>transition"`
	Result  *Result       `xml:"result"`
	Stats   *Statistics   `xml:"statistics"`
	Success bool
}


func (this *JPFReport) String() string {
	return fmt.Sprintf("Id: %q; Version: %s; \nResult: %s;\n Stats: %s",
		this.Id, this.Version, this.Result, this.Stats)
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
	CG       *ChoiceGenerator `xml:"cg"`
	Insns    []*Instruction   `xml:"insn"`
}

func (this *Transition) String() string {
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
	Errors   []*Error `xml:"error"`
}

func (this *Result) String() string {
	return fmt.Sprintf("Findings: %s", this.Findings)
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


func (this *Statistics) String() string {
	return fmt.Sprintf("NewStates: %d; VisitedStates: %d; BacktrackedStates: %d; EndStates: %d;", 
		this.NewStates, this.VisitedStates, this.BacktrackedStates, this.EndStates)
}

func genReport(id bson.ObjectId, data []byte) (res *JPFReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		err = tool.NewXMLError(err, "jpf/jpfResult.go")
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
