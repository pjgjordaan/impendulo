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
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/tool/wc"

	"labix.org/v2/mgo/bson"

	"strings"
)

type (
	//C is a Displayer used to display a source file's code.
	C struct {
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
func New(fid bson.ObjectId, lang project.Language, data []byte) *C {
	return &C{
		FileId: fid,
		Lang:   lang,
		Data:   strings.TrimSpace(string(data)),
	}
}

func (c *C) GetType() string {
	return NAME
}

//GetName
func (c *C) GetName() string {
	return c.GetType()
}

//Reporter
func (c *C) Reporter() result.Reporter {
	return c
}

func (c *C) Template() string {
	return "coderesult"
}

func (c *C) ChartVals() []*result.ChartVal {
	return []*result.ChartVal{{Name: "Lines", Y: float64(c.Lines()), FileId: c.FileId}}
}

func (c *C) Lines() int64 {
	lc, _ := wc.Lines(c.Data)
	return lc
}

func (c *C) ChartVal(n string) (*result.ChartVal, error) {
	switch n {
	case LINES:
		return &result.ChartVal{Name: LINES, Y: float64(c.Lines()), FileId: c.FileId}, nil
	default:
		return nil, fmt.Errorf("unknown ChartVal %s", n)
	}
}

func Types() []string {
	return []string{LINES}
}

func (c *C) GetFileId() bson.ObjectId {
	return c.FileId
}
