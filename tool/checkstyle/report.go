//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package checkstyle

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
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
		Name   string   `xml:"name,attr"`
		Errors []*Error `xml:"error"`
	}

	//Problem is a compressed variation of Error.
	//This is because it stores all errors of the same type in a single struct.
	//This is done by storing the lines in which they occur in a seperate array.
	Problem struct {
		*Error
		Lines []int
	}

	//Error represents an occurrence of an error detected by checkstyle.
	//It gives the location of the error, its severity and a thorough description.
	Error struct {
		Line     int           `xml:"line,attr"`
		Column   int           `xml:"column,attr"`
		Severity string        `xml:"severity,attr"`
		Message  template.HTML `xml:"message,attr"`
		Source   string        `xml:"source,attr"`
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
	}
	return
}

//File
func (this *Report) File(name string) *File {
	for _, f := range this.Files {
		if strings.HasSuffix(f.Name, name) {
			return f
		}
	}
	return nil
}

//Success
func (this *Report) Success() bool {
	return this.Errors == 0
}

//String
func (this *Report) String() string {
	files := ""
	for _, f := range this.Files {
		files += f.String()
	}
	return fmt.Sprintf("Id: %q; Version %s; Errors: %d; \nFiles: %s\n",
		this.Id, this.Version, this.Errors, files)
}

//ShouldDisplay
func (this *File) ShouldDisplay() bool {
	return len(this.Errors) > 0
}

//String
func (this *File) String() string {
	errs := ""
	for _, e := range this.Errors {
		errs += e.String()
	}
	return fmt.Sprintf("Name: %s; \nErrors: %s\n",
		this.Name, errs)
}

//Problems
func (this *File) Problems() map[string]*Problem {
	problems := make(map[string]*Problem)
	for _, e := range this.Errors {
		p, ok := problems[e.Source]
		if !ok {
			problems[e.Source] = &Problem{e, make([]int, 0, len(this.Errors))}
			p = problems[e.Source]
		}
		p.Lines = append(p.Lines, e.Line)
	}
	return problems
}

//String
func (this *Error) String() string {
	return fmt.Sprintf("Line: %d; Column: %d; Severity: %s; Message: %q; Source: %s\n",
		this.Line, this.Column, this.Severity, this.Message, this.Source)
}
