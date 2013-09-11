package pmd

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
	"math"
)

const NAME = "PMD"

func init() {
	gob.Register(new(Report))
}

type Result struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	Data   *Report       "data"
	GridFS bool          "gridfs"
}

func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

func (this *Result) OnGridFS() bool {
	return this.GridFS
}

func (this *Result) GetName() string {
	return this.Name
}

func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Result) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d", this.Data.Errors)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *Result) GetData() interface{} {
	return this.Data
}

func (this *Result) Template(current bool) string {
	if current {
		return "pmdCurrent"
	} else {
		return "pmdNext"
	}
}

func (this *Result) Success() bool {
	return true
}

func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("PMD Errors")
	}
	y := float64(this.Data.Errors)
	tool.AddCoords(graphData[0], x, y)
	return math.Max(max, y)
}

func (this *Result) String() (ret string) {
	if this.Data != nil {
		ret = this.Data.String()
	}
	ret += this.Id.Hex()
	return
}

func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
	}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *Report, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		err = tool.NewXMLError(err, "pmd/pmdResult.go")
		return
	}
	res.Id = id
	res.Errors = 0
	for _, f := range res.Files {
		res.Errors += len(f.Violations)
	}
	return
}

type Report struct {
	Id      bson.ObjectId
	Version string  `xml:"version,attr"`
	Files   []*File `xml:"file"`
	Errors  int
}

func (this *Report) Success() bool {
	return this.Errors == 0
}

func (this *Report) String() (ret string) {
	ret = fmt.Sprintf("Report{ Errors: %d\n.", this.Errors)
	if this.Files != nil {
		ret += "Files: \n"
		for _, f := range this.Files {
			ret += f.String()
		}
	}
	ret += "}\n"
	return
}

type File struct {
	Name       string       `xml:"name,attr"`
	Violations []*Violation `xml:"violation"`
}

func (this *File) Problems() map[string]*Problem {
	problems := make(map[string]*Problem)
	for _, v := range this.Violations {
		p, ok := problems[v.Rule]
		if !ok {
			problems[v.Rule] = &Problem{v,
				make([]int, 0, len(this.Violations)), make([]int, 0, len(this.Violations))}
			p = problems[v.Rule]
		}
		p.Starts = append(p.Starts, v.Begin)
		p.Ends = append(p.Ends, v.End)
	}
	return problems
}

type Problem struct {
	*Violation
	Starts, Ends []int
}

func (this *File) String() (ret string) {
	ret = fmt.Sprintf("File{ Name: %s\n.", this.Name)
	if this.Violations != nil {
		ret += "Violations: \n"
		for _, v := range this.Violations {
			ret += v.String()
		}
	}
	ret += "}\n"
	return
}

type Violation struct {
	Begin       int          `xml:"beginline,attr"`
	End         int          `xml:"endline,attr"`
	Rule        string       `xml:"rule,attr"`
	RuleSet     string       `xml:"ruleset,attr"`
	Url         template.URL `xml:"externalInfoUrl,attr"`
	Priority    int          `xml:"priority,attr"`
	Description string       `xml:",innerxml"`
}

func (this *Violation) String() (ret string) {
	ret = fmt.Sprintf("Violation{ Begin: %d; End: %d; Rule: %s; RuleSet: %s; "+
		"Priority: %d; Description: %s}\n",
		this.Begin, this.End, this.Rule, this.RuleSet,
		this.Priority, this.Description)
	return
}
