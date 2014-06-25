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

package checkstyle

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"

	"html/template"

	"labix.org/v2/mgo/bson"

	"sort"
	"strings"
)

type (
	//Report represents the result of running Checkstyle on a Java source file.
	Report struct {
		Id      bson.ObjectId
		Version string `xml:"version,attr"`
		Errors  int
		Files   []*File `xml:"file"`
	}

	//File represents a file on which checkstyle was run and all errors found in it.
	File struct {
		Name   string `xml:"name,attr"`
		Errors Errors `xml:"error"`
	}

	//Errors
	Errors []*Error

	//Error represents an occurrence of an error detected by checkstyle.
	//It gives the location of the error, its severity and a thorough description.
	Error struct {
		Id       bson.ObjectId
		Line     int           `xml:"line,attr"`
		Column   int           `xml:"column,attr"`
		Severity string        `xml:"severity,attr"`
		Message  template.HTML `xml:"message,attr"`
		Source   string        `xml:"source,attr"`
		Lines    []int
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (res *Report, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		err = tool.NewXMLError(err, "checkstyle/checkstyleResult.go")
		return
	}
	res.Id = id
	res.Errors = 0
	for _, f := range res.Files {
		res.Errors += len(f.Errors)
		f.CompressErrors()
	}
	return
}

func (r *Report) Lines() []*result.Line {
	lines := make([]*result.Line, 0, len(r.Files)*10)
	for _, f := range r.Files {
		for _, e := range f.Errors {
			for _, l := range e.Lines {
				lines = append(lines, &result.Line{Title: util.ShortName(e.Source), Description: string(e.Message), Start: l, End: l})
			}
		}
	}
	return lines
}

//File
func (r *Report) File(name string) *File {
	for _, f := range r.Files {
		if strings.HasSuffix(f.Name, name) {
			return f
		}
	}
	return nil
}

//Success
func (r *Report) Success() bool {
	return r.Errors == 0
}

//String
func (r *Report) String() string {
	files := ""
	for _, f := range r.Files {
		files += f.String()
	}
	return fmt.Sprintf("Id: %q; Version %s; Errors: %d; \nFiles: %s\n",
		r.Id, r.Version, r.Errors, files)
}

//String
func (f *File) String() string {
	errs := ""
	for _, e := range f.Errors {
		errs += e.String()
	}
	return fmt.Sprintf("Name: %s; \nErrors: %s\n",
		f.Name, errs)
}

//CompressErrors packs all Errors of the same
//type into a single Error by storing their location seperately.
func (f *File) CompressErrors() {
	indices := make(map[string]int)
	compressed := make(Errors, 0, len(f.Errors))
	for _, e := range f.Errors {
		index, ok := indices[e.Source]
		if !ok {
			e.Lines = make([]int, 0, len(f.Errors))
			e.Id = bson.NewObjectId()
			compressed = append(compressed, e)
			index = len(compressed) - 1
			indices[e.Source] = index
		}
		compressed[index].Lines = append(compressed[index].Lines, e.Line)
	}
	sort.Sort(compressed)
	f.Errors = compressed
}

func (e Errors) Len() int {
	return len(e)
}

func (e Errors) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e Errors) Less(i, j int) bool {
	return e[i].Source < e[j].Source
}

//String
func (e *Error) String() string {
	return fmt.Sprintf("Line: %d; Column: %d; Severity: %s; Message: %q; Source: %s\n",
		e.Line, e.Column, e.Severity, e.Message, e.Source)
}
