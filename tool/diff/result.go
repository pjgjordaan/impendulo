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

//Package diff is used to run diffs on source files and provide the result in HTML.
package diff

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"

	"strings"
)

const (
	NAME = "Diff"
)

type (
	//Result is a tool.DisplayResult used to display a diff between two files.
	Result struct {
		header, data string
	}
)

//NewResult creates a new Result with a single file.
//A Result is actually only made up of a single file's source code
//and never contains a diff. This is calculated seperately.
func NewResult(f *project.File) *Result {
	return &Result{
		header: f.Name + " " + util.Date(f.Time),
		data:   strings.TrimSpace(string(f.Data)),
	}
}

//Create calculates the diff of two Results' code, converts it to HTML
//and returns this.
func (r *Result) Create(next *Result) (string, error) {
	d, e := Diff(r.data, next.data)
	if e != nil {
		return "", e
	}
	return SetHeader(d, r.header, next.header), nil
}

func (r *Result) GetType() string {
	return NAME
}

//GetName
func (r *Result) GetName() string {
	return NAME
}

//GetReport
func (r *Result) GetReport() tool.Report {
	return r
}

func (r *Result) Template() string {
	return "diffresult"
}
