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

package pmd

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
)

type (
	//Report is the result of running PMD on a Java source file.
	Report struct {
		Id      bson.ObjectId
		Version string  `xml:"version,attr"`
		Files   []*File `xml:"file"`
		Errors  int
	}

	//File
	File struct {
		Name       string       `xml:"name,attr"`
		Violations []*Violation `xml:"violation"`
	}

	//Problem
	Problem struct {
		*Violation
		Starts, Ends []int
	}

	//Violation
	Violation struct {
		Begin       int          `xml:"beginline,attr"`
		End         int          `xml:"endline,attr"`
		Rule        string       `xml:"rule,attr"`
		RuleSet     string       `xml:"ruleset,attr"`
		Url         template.URL `xml:"externalInfoUrl,attr"`
		Priority    int          `xml:"priority,attr"`
		Description string       `xml:",innerxml"`
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (res *Report, err error) {
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

//Success
func (this *Report) Success() bool {
	return this.Errors == 0
}

//String
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

//Problems
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

//String
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

//String
func (this *Violation) String() (ret string) {
	ret = fmt.Sprintf("Violation{ Begin: %d; End: %d; Rule: %s; RuleSet: %s; "+
		"Priority: %d; Description: %s}\n",
		this.Begin, this.End, this.Rule, this.RuleSet,
		this.Priority, this.Description)
	return
}
