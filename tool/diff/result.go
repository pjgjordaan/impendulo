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
