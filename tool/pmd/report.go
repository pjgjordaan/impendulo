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
	Report struct {
		Id      bson.ObjectId
		Version string  `xml:"version,attr"`
		Files   []*File `xml:"file"`
		Errors  int
	}

	File struct {
		Name       string       `xml:"name,attr"`
		Violations []*Violation `xml:"violation"`
	}

	Problem struct {
		*Violation
		Starts, Ends []int
	}

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

func (this *Report) Success() bool {
	return this.Errors == 0
}

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

func (this *Violation) String() (ret string) {
	ret = fmt.Sprintf("Violation{ Begin: %d; End: %d; Rule: %s; RuleSet: %s; "+
		"Priority: %d; Description: %s}\n",
		this.Begin, this.End, this.Rule, this.RuleSet,
		this.Priority, this.Description)
	return
}
