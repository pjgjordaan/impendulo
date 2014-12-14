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

package code

import (
	"fmt"

	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/lc"
	"github.com/godfried/impendulo/tool/result"

	"labix.org/v2/mgo/bson"

	"strings"
)

type (
	//Result is a Displayer used to display a source file's code.
	Result struct {
		FileId bson.ObjectId
		Lang   project.Language
		Data   string
	}
)

const (
	NAME  = "Code"
	LINES = "Lines"
)

//New
func New(fid bson.ObjectId, lang project.Language, data []byte) *Result {
	return &Result{
		FileId: fid,
		Lang:   lang,
		Data:   strings.TrimSpace(string(data)),
	}
}

func (r *Result) GetType() string {
	return NAME
}

//GetName
func (r *Result) GetName() string {
	return r.GetType()
}

//Reporter
func (r *Result) Reporter() result.Reporter {
	return r
}

func (r *Result) Template() string {
	return "coderesult"
}

func (r *Result) Values() []*result.Value {
	return []*result.Value{{Name: "Lines", V: float64(r.Lines()), FileId: r.FileId}}
}

func (r *Result) Lines() int64 {
	lc, _ := lc.Lines(r.Data)
	return lc
}

func (r *Result) Value(n string) (*result.Value, error) {
	switch n {
	case LINES:
		return &result.Value{Name: LINES, V: float64(r.Lines()), FileId: r.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown Value %s", n)
	}
}

func (r *Result) Types() []string {
	return []string{LINES}
}

func (r *Result) GetFileId() bson.ObjectId {
	return r.FileId
}
