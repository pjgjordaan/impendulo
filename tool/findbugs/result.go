package findbugs

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
	"math"
)

const NAME = "Findbugs"

func init() {
	gob.Register(new(Report))
}

//FindBugsResult is a tool.ToolResult and a tool.DisplayResult.
//It is used to store the output of running Findbugs.
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

func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Data: %s",
		this.Id, this.FileId, this.Name, this.Data)
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
	body := fmt.Sprintf("Bugs: %d", this.Data.Summary.BugCount)
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
		return "findbugsCurrent"
	} else {
		return "findbugsNext"
	}
}

func (this *Result) Success() bool {
	return true
}

func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("Findbugs All")
		graphData[1] = tool.CreateChart("Findbugs Priority 1")
		graphData[2] = tool.CreateChart("Findbugs Priority 2")
		graphData[3] = tool.CreateChart("Findbugs Priority 3")
	}
	yB := float64(this.Data.Summary.BugCount)
	y1 := float64(this.Data.Summary.Priority1)
	y2 := float64(this.Data.Summary.Priority2)
	y3 := float64(this.Data.Summary.Priority3)
	tool.AddCoords(graphData[0], x, yB)
	tool.AddCoords(graphData[1], x, y1)
	tool.AddCoords(graphData[2], x, y2)
	tool.AddCoords(graphData[3], x, y3)
	return math.Max(max, math.Max(y1, math.Max(y2, math.Max(y3, yB))))
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
		err = tool.NewXMLError(err, "findbugs/findbugsResult.go")
		return
	}
	res.loadMaps()
	res.Id = id
	return
}

//Report stores the results of running Findbugs. It is
//populated from XML output produced by findbugs.
type Report struct {
	Id          bson.ObjectId
	Time        int            `xml:"analysisTimestamp,attr"`
	Summary     *Summary       `xml:"FindBugsSummary"`
	Instances   []*BugInstance `xml:"BugInstance"`
	Categories  []*BugCategory `xml:"BugCategory"`
	Patterns    []*BugPattern  `xml:"BugPattern"`
	CategoryMap map[string]*BugCategory
	PatternMap  map[string]*BugPattern
}

func (this *Report) Success() bool {
	return len(this.Instances) == 0
}

func (this *Report) String() string {
	return fmt.Sprintf("Id: %q; Summary: %s",
		this.Id, this.Summary)
}

func (this *Report) loadMaps() {
	this.CategoryMap = make(map[string]*BugCategory)
	this.PatternMap = make(map[string]*BugPattern)
	for _, cat := range this.Categories {
		this.CategoryMap[cat.Name] = cat
	}
	for _, pat := range this.Patterns {
		this.PatternMap[pat.Type] = pat
	}
}

type Summary struct {
	ClassCount     int             `xml:"total_classes,attr"`
	ReferenceCount int             `xml:"referenced_classes,attr"`
	BugCount       int             `xml:"total_bugs,attr"`
	Size           int             `xml:"total_size,attr"`
	PackageCount   int             `xml:"num_packages,attr"`
	SecondsCPU     int             `xml:"cpu_seconds,attr"`
	SecondsClock   int             `xml:"clock_seconds,attr"`
	SecondsGC      int             `xml:"gc_seconds,attr"`
	PeakMB         int             `xml:"peak_mbytes,attr"`
	AllocMB        int             `xml:"alloc_mbytes,attr"`
	Priority1      int             `xml:"priority_1,attr"`
	Priority2      int             `xml:"priority_2,attr"`
	Priority3      int             `xml:"priority_3,attr"`
	Files          []*FileStats    `xml:"FileStats"`
	Packages       []*PackageStats `xml:"PackageStats"`
}

func (this *Summary) String() string {
	return fmt.Sprintf("BugCount: %d; ClassCount: %d",
		this.BugCount, this.ClassCount)
}

type FileStats struct {
	Path     string `xml:"path,attr"`
	BugCount int    `xml:"bugCount,attr"`
	Size     int    `xml:"size,attr"`
}

type PackageStats struct {
	Name       string        `xml:"package,attr"`
	ClassCount int           `xml:"total_types,attr"`
	BugCount   int           `xml:"total_bugs,attr"`
	Size       int           `xml:"total_size,attr"`
	Priority1  int           `xml:"priority_1,attr"`
	Priority2  int           `xml:"priority_2,attr"`
	Priority3  int           `xml:"priority_3,attr"`
	Classes    []*ClassStats `xml:"ClassStats"`
}

type ClassStats struct {
	Name        string `xml:"class,attr"`
	Source      string `xml:"sourceFile,attr"`
	IsInterface bool   `xml:"interface,attr"`
	BugCount    int    `xml:"bugs,attr"`
	Size        int    `xml:"size,attr"`
	Priority1   int    `xml:"priority_1,attr"`
	Priority2   int    `xml:"priority_2,attr"`
	Priority3   int    `xml:"priority_3,attr"`
}

type BugInstance struct {
	Type         string         `xml:"type,attr"`
	Priority     int            `xml:"priority,attr"`
	Abbreviation string         `xml:"abbrev,attr"`
	Category     string         `xml:"category,attr"`
	Rank         int            `xml:"rank,attr"`
	ShortMessage string         `xml:"ShortMessage"`
	LongMessage  string         `xml:"LongMessage"`
	Class        *Class         `xml:"Class"`
	Method       *Method        `xml:"Method"`
	Field        *Field         `xml:"Field"`
	Var          *LocalVariable `xml:"LocalVariable"`
	Line         *SourceLine    `xml:"SourceLine"`
	Properties   []*Property    `xml:"Property"`
}

type Class struct {
	Name      string      `xml:"classname,attr"`
	IsPrimary bool        `xml:"primary,attr"`
	Line      *SourceLine `xml:"SourceLine"`
	Message   string      `xml:"Message"`
}

type SourceLine struct {
	Class   string `xml:"classname,attr"`
	Start   int    `xml:"start,attr"`
	End     int    `xml:"end,attr"`
	StartBC int    `xml:"startBytecode,attr"`
	EndBC   int    `xml:"endBytecode,attr"`
	File    string `xml:"sourcefile,attr"`
	Path    string `xml:"sourcepath,attr"`
	Message string `xml:"Message"`
}

type Method struct {
	Name      string      `xml:"name,attr"`
	Class     string      `xml:"classname,attr"`
	Signature string      `xml:"signature,attr"`
	IsStatic  bool        `xml:"isStatic,attr"`
	IsPrimary bool        `xml:"primary,attr"`
	Line      *SourceLine `xml:"SourceLine"`
	Message   string      `xml:"Message"`
}

type Field struct {
	Name      string      `xml:"name,attr"`
	Class     string      `xml:"classname,attr"`
	Signature string      `xml:"signature,attr"`
	IsStatic  bool        `xml:"isStatic,attr"`
	IsPrimary bool        `xml:"primary,attr"`
	Role      string      `xml:"role,attr"`
	Line      *SourceLine `xml:"SourceLine"`
	Message   string      `xml:"Message"`
}

type LocalVariable struct {
	Name     string `xml:"name,attr"`
	Register int    `xml:"register,attr"`
	PC       string `xml:"pc,attr"`
	Role     string `xml:"role,attr"`
	Message  string `xml:"Message"`
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type BugCategory struct {
	Name        string `xml:"category,attr"`
	Description string `xml:"Description"`
}

type BugPattern struct {
	Type         string        `xml:"type,attr"`
	Abbreviation string        `xml:"abbrev,attr"`
	Category     string        `xml:"category,attr"`
	Description  string        `xml:"ShortDescription"`
	Details      template.HTML `xml:"Details"`
}
