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

//Package diff is used to run diffs on source files and provide the result in HTML.
package diff

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"html/template"
	"strings"
)

const (
	NAME = "Diff"
)

type (
	//DiffResult is a DisplayResult used to display a diff between two files.
	DiffResult struct {
		header, data string
	}
)

func NewDiffResult(file *project.File) *DiffResult {
	header := file.Name + " " + util.Date(file.Time)
	data := strings.TrimSpace(string(file.Data))
	return &DiffResult{
		header: header,
		data:   data,
	}
}

func (this *DiffResult) Create(next *DiffResult) (ret template.HTML, err error) {
	diff, err := Diff(this.data, next.data)
	if err != nil {
		return
	}
	diff = SetHeader(diff, this.header, next.header)
	ret, err = Diff2HTML(diff)
	return

}

func (this *DiffResult) GetName() string {
	return NAME
}

func (this *DiffResult) GetData() interface{} {
	return this
}
