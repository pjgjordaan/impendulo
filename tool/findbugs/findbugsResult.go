package findbugs

import (
	"encoding/xml"
	"github.com/godfried/impendulo/util"
	"html/template"
	"labix.org/v2/mgo/bson"
)

const NAME = "Findbugs"

type FindbugsResult struct {
	Id     bson.ObjectId   "_id"
	FileId bson.ObjectId   "fileid"
	Time   int64           "time"
	Data   *FindbugsReport "data"
}

func (this *FindbugsResult) GetName() string {
	return NAME
}

func (this *FindbugsResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *FindbugsResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *FindbugsResult) String() string {
	return "Type: tool.java.FindbugsResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *FindbugsResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "findbugsCurrent", this.Data
	} else {
		return "findbugsNext", this.Data
	}
}

func (this *FindbugsResult) Success() bool {
	return true
}

func NewResult(fileId bson.ObjectId, data []byte) (res *FindbugsResult, err error) {
	res = &FindbugsResult{Id: bson.NewObjectId(), FileId: fileId, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *FindbugsReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		return
	}
	res.loadMaps()
	res.Id = id
	return
}

type FindbugsReport struct {
	Id          bson.ObjectId
	Time        int64            `xml:"analysisTimestamp,attr"`
	Summary     *FindBugsSummary `xml:"FindBugsSummary"`
	Instances   []*BugInstance   `xml:"BugInstance"`
	Categories  []*BugCategory   `xml:"BugCategory"`
	Patterns    []*BugPattern    `xml:"BugPattern"`
	CategoryMap map[string]*BugCategory
	PatternMap  map[string]*BugPattern
}

func (this *FindbugsReport) loadMaps() {
	this.CategoryMap = make(map[string]*BugCategory)
	this.PatternMap = make(map[string]*BugPattern)
	for _, cat := range this.Categories {
		this.CategoryMap[cat.Name] = cat
	}
	for _, pat := range this.Patterns {
		this.PatternMap[pat.Type] = pat
	}
}

type FindBugsSummary struct {
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
	Type         string        `xml:"type,attr"`
	Priority     int           `xml:"priority,attr"`
	Abbreviation string        `xml:"abbrev,attr"`
	Category     string        `xml:"category,attr"`
	Rank         int           `xml:"rank,attr"`
	ShortMessage string        `xml:"ShortMessage"`
	LongMessage  string        `xml:"LongMessage"`
	Class        Class         `xml:"Class"`
	Method       Method        `xml:"Method"`
	Field        Field         `xml:"Field"`
	Var          LocalVariable `xml:"LocalVariable"`
	Line         SourceLine    `xml:"SourceLine"`
	Properties   []*Property   `xml:"Property"`
}

type Class struct {
	Name      string     `xml:"classname,attr"`
	IsPrimary bool       `xml:"primary,attr"`
	Line      SourceLine `xml:"SourceLine"`
	Message   string     `xml:"Message"`
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
	Name      string     `xml:"name,attr"`
	Class     string     `xml:"classname,attr"`
	Signature string     `xml:"signature,attr"`
	IsStatic  bool       `xml:"isStatic,attr"`
	IsPrimary bool       `xml:"primary,attr"`
	Line      SourceLine `xml:"SourceLine"`
	Message   string     `xml:"Message"`
}

type Field struct {
	Name      string     `xml:"name,attr"`
	Class     string     `xml:"classname,attr"`
	Signature string     `xml:"signature,attr"`
	IsStatic  bool       `xml:"isStatic,attr"`
	IsPrimary bool       `xml:"primary,attr"`
	Role      string     `xml:"role,attr"`
	Line      SourceLine `xml:"SourceLine"`
	Message   string     `xml:"Message"`
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
