package findbugs

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
)

type (
	//Report stores the results of running Findbugs. It is
	//populated from XML output produced by findbugs.
	Report struct {
		Id      bson.ObjectId
		Time    int      `xml:"analysisTimestamp,attr"`
		Summary *Summary `xml:"FindBugsSummary"`
		//Instances is all the bugs found by Findbugs
		Instances []*BugInstance `xml:"BugInstance"`
		//Categories is the bug categories found by Findbugs.
		Categories []*BugCategory `xml:"BugCategory"`
		//Patterns is the bug patterns found by Findbugs.
		Patterns []*BugPattern `xml:"BugPattern"`
		//CategoryMap and PatternMap make it easier to use the bug categories and patterns.
		CategoryMap map[string]*BugCategory
		PatternMap  map[string]*BugPattern
	}

	//Summary provides statistics about Findbugs's execution on a package and
	//file level. Furthermore it gives performance information such as memory usage
	//and where time was spent.
	Summary struct {
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

	//FileStats provides Findbugs statistics about a specific source file.
	FileStats struct {
		Path     string `xml:"path,attr"`
		BugCount int    `xml:"bugCount,attr"`
		Size     int    `xml:"size,attr"`
	}

	//PackageStats provides Findbugs statistics about a specific package
	//as well as the classes found within it.
	PackageStats struct {
		Name       string        `xml:"package,attr"`
		ClassCount int           `xml:"total_types,attr"`
		BugCount   int           `xml:"total_bugs,attr"`
		Size       int           `xml:"total_size,attr"`
		Priority1  int           `xml:"priority_1,attr"`
		Priority2  int           `xml:"priority_2,attr"`
		Priority3  int           `xml:"priority_3,attr"`
		Classes    []*ClassStats `xml:"ClassStats"`
	}

	//ClassStats provides Findbugs statistics about a specific class
	//within a package.
	ClassStats struct {
		Name        string `xml:"class,attr"`
		Source      string `xml:"sourceFile,attr"`
		IsInterface bool   `xml:"interface,attr"`
		BugCount    int    `xml:"bugs,attr"`
		Size        int    `xml:"size,attr"`
		Priority1   int    `xml:"priority_1,attr"`
		Priority2   int    `xml:"priority_2,attr"`
		Priority3   int    `xml:"priority_3,attr"`
	}

	//BugInstance describes a particular bug detected by Findbugs.
	//It contains information describing its location and state
	//as well as its category, severity and type.
	BugInstance struct {
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

	//Class describes a Java Class in which a bug was found.
	Class struct {
		Name      string      `xml:"classname,attr"`
		IsPrimary bool        `xml:"primary,attr"`
		Line      *SourceLine `xml:"SourceLine"`
		Message   string      `xml:"Message"`
	}

	//Sourceline describes a line inside a Java Class.
	SourceLine struct {
		Class   string `xml:"classname,attr"`
		Start   int    `xml:"start,attr"`
		End     int    `xml:"end,attr"`
		StartBC int    `xml:"startBytecode,attr"`
		EndBC   int    `xml:"endBytecode,attr"`
		File    string `xml:"sourcefile,attr"`
		Path    string `xml:"sourcepath,attr"`
		Message string `xml:"Message"`
	}

	//Method describes a method within a Java Class.
	Method struct {
		Name      string      `xml:"name,attr"`
		Class     string      `xml:"classname,attr"`
		Signature string      `xml:"signature,attr"`
		IsStatic  bool        `xml:"isStatic,attr"`
		IsPrimary bool        `xml:"primary,attr"`
		Line      *SourceLine `xml:"SourceLine"`
		Message   string      `xml:"Message"`
	}

	//Field describes a global variable within a Java Class.
	Field struct {
		Name      string      `xml:"name,attr"`
		Class     string      `xml:"classname,attr"`
		Signature string      `xml:"signature,attr"`
		IsStatic  bool        `xml:"isStatic,attr"`
		IsPrimary bool        `xml:"primary,attr"`
		Role      string      `xml:"role,attr"`
		Line      *SourceLine `xml:"SourceLine"`
		Message   string      `xml:"Message"`
	}

	//LocalVariable describes a local variable within a Java Class.
	LocalVariable struct {
		Name     string `xml:"name,attr"`
		Register int    `xml:"register,attr"`
		PC       string `xml:"pc,attr"`
		Role     string `xml:"role,attr"`
		Message  string `xml:"Message"`
	}

	//Property decribes some attribute associated with a bug.
	Property struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr"`
	}

	//BugCategory describes a category in which bugs may fall.
	BugCategory struct {
		Name        string `xml:"category,attr"`
		Description string `xml:"Description"`
	}

	//BugPattern describes a pattern associated with a BugCategory.
	BugPattern struct {
		Type         string        `xml:"type,attr"`
		Abbreviation string        `xml:"abbrev,attr"`
		Category     string        `xml:"category,attr"`
		Description  string        `xml:"ShortDescription"`
		Details      template.HTML `xml:"Details"`
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (res *Report, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		err = tool.NewXMLError(err, "findbugs/findbugsResult.go")
		return
	}
	res.loadMaps()
	res.Id = id
	return
}

//Success returns true if Findbugs found no bugs and false otherwise.
func (this *Report) Success() bool {
	return len(this.Instances) == 0
}

//String
func (this *Report) String() string {
	return fmt.Sprintf("Id: %q; Summary: %s",
		this.Id, this.Summary)
}

//loadMaps generates maps from Categories and Paterns for easier access.
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

//String
func (this *Summary) String() string {
	return fmt.Sprintf("BugCount: %d; ClassCount: %d",
		this.BugCount, this.ClassCount)
}
