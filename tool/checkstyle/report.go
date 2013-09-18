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
	Report struct {
		Id      bson.ObjectId
		Version string `xml:"version,attr"`
		Errors  int
		Files   []*File `xml:"file"`
	}

	File struct {
		Name   string   `xml:"name,attr"`
		Errors []*Error `xml:"error"`
	}

	Problem struct {
		*Error
		Lines []int
	}

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

func (this *Report) File(name string) *File {
	for _, f := range this.Files {
		if strings.HasSuffix(f.Name, name) {
			return f
		}
	}
	return nil
}

func (this *Report) Success() bool {
	return this.Errors == 0
}

func (this *Report) String() string {
	files := ""
	for _, f := range this.Files {
		files += f.String()
	}
	return fmt.Sprintf("Id: %q; Version %s; Errors: %d; \nFiles: %s\n",
		this.Id, this.Version, this.Errors, files)
}

func (this *File) ShouldDisplay() bool {
	return len(this.Errors) > 0
}

func (this *File) String() string {
	errs := ""
	for _, e := range this.Errors {
		errs += e.String()
	}
	return fmt.Sprintf("Name: %s; \nErrors: %s\n",
		this.Name, errs)
}

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

func (this *Error) String() string {
	return fmt.Sprintf("Line: %d; Column: %d; Severity: %s; Message: %q; Source: %s\n",
		this.Line, this.Column, this.Severity, this.Message, this.Source)
}
