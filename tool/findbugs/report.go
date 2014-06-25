//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package findbugs

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"

	"html/template"

	"labix.org/v2/mgo/bson"
)

type (
	DummyReport struct {
		Time    int      `xml:"analysisTimestamp,attr"`
		Summary *Summary `xml:"FindBugsSummary"`
		//Instances is all the bugs found by Findbugs
		Instances []*BugInstance `xml:"BugInstance"`
		//Categories is the bug categories found by Findbugs.
		Categories []*BugCategory `xml:"BugCategory"`
		//Patterns is the bug patterns found by Findbugs.
		Patterns []*BugPattern `xml:"BugPattern"`
	}

	//Report stores the results of running Findbugs. It is
	//populated from XML output produced by findbugs.
	Report struct {
		Id      bson.ObjectId
		Time    int
		Summary *Summary
		//Instances is all the bugs found by Findbugs
		Instances []*BugInstance
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
		Id           bson.ObjectId
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
func NewReport(id bson.ObjectId, data []byte) (*Report, error) {
	var d *DummyReport
	if e := xml.Unmarshal(data, &d); e != nil {
		return nil, tool.NewXMLError(e, "findbugs/findbugsResult.go")
	}
	r := &Report{
		Id:          id,
		Time:        d.Time,
		Summary:     d.Summary,
		Instances:   d.Instances,
		CategoryMap: make(map[string]*BugCategory),
		PatternMap:  make(map[string]*BugPattern),
	}
	for _, c := range d.Categories {
		r.CategoryMap[c.Name] = c
	}
	for _, p := range d.Patterns {
		r.PatternMap[p.Type] = p
	}
	for _, b := range r.Instances {
		b.Id = bson.NewObjectId()
	}
	return r, nil
}

//Success returns true if Findbugs found no bugs and false otherwise.
func (r *Report) Success() bool {
	return len(r.Instances) == 0
}

//String
func (r *Report) String() string {
	return fmt.Sprintf("Id: %q; Summary: %s", r.Id, r.Summary)
}

func (r *Report) Lines() []*result.Line {
	lines := make([]*result.Line, 0, len(r.Instances))
	for _, b := range r.Instances {
		lines = append(lines, &result.Line{Title: r.PatternMap[b.Type].Description, Description: b.LongMessage, Start: b.Line.Start, End: b.Line.End})
	}
	return lines
}

//String
func (s *Summary) String() string {
	return fmt.Sprintf("BugCount: %d; ClassCount: %d", s.BugCount, s.ClassCount)
}
