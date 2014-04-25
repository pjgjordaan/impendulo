package jacoco

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"strings"
)

type (
	Report struct {
		Id           bson.ObjectId
		Name         string     `xml:"name,attr"`
		Packages     []*Package `xml:"package"`
		Counters     []*Counter `xml:"counter"`
		HTML         []byte
		MainCounters []*Counter
	}

	Package struct {
		Name     string     `xml:"name,attr"`
		Classes  []*Class   `xml:"class"`
		Sources  []*Source  `xml:"sourcefile"`
		Counters []*Counter `xml:"counter"`
	}

	Class struct {
		Name     string     `xml:"name,attr"`
		Methods  []*Method  `xml:"method"`
		Counters []*Counter `xml:"counter"`
	}

	Method struct {
		Name     string     `xml:"name,attr"`
		Desc     string     `xml:"desc,attr"`
		Line     uint       `xml:"line,attr"`
		Counters []*Counter `xml:"counter"`
	}

	Counter struct {
		Type    string `xml:"type,attr"`
		Missed  uint   `xml:"missed,attr"`
		Covered uint   `xml:"covered,attr"`
	}
	Source struct {
		Name     string     `xml:"name,attr"`
		Lines    []*Line    `xml:"line"`
		Counters []*Counter `xml:"counter"`
	}
	Line struct {
		NR uint `xml:"nr,attr"`
		MI uint `xml:"mi,attr"`
		CI uint `xml:"ci,attr"`
		MB uint `xml:"mb,attr"`
		CB uint `xml:"cb,attr"`
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, xmlData, htmlData []byte, target *tool.Target) (*Report, error) {
	var r *Report
	if e := xml.Unmarshal(xmlData, &r); e != nil {
		if r == nil {
			return nil, tool.NewXMLError(e, "jacoco/jacocoResult.go")
		} else {
			return nil, e
		}
	}
	r.HTML = htmlData
	r.Id = id
	for _, p := range r.Packages {
		if p.Name != target.Package {
			continue
		}
		for _, cl := range p.Classes {
			if !strings.HasSuffix(cl.Name, target.Name) {
				continue
			}
			r.MainCounters = cl.Counters
			return r, nil
		}
	}
	return nil, fmt.Errorf("no class matching %s found", target.Executable())
}

func (r *Report) Class(name string) (*Class, error) {
	n, _ := util.Extension(name)
	for _, p := range r.Packages {
		for _, c := range p.Classes {
			if strings.HasSuffix(c.Name, name) || strings.HasSuffix(c.Name, n) {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("no matching class found for %s", name)
}

func (r *Report) String() string {
	ps := "{\n"
	for _, p := range r.Packages {
		ps += p.String()
	}
	ps += "}"
	cs := "{\n"
	for _, c := range r.Counters {
		cs += c.String()
	}
	cs += "}"
	return fmt.Sprintf("name %s\npackages %s\ncounters %s\n", r.Name, ps, cs)
}

func (p *Package) String() string {
	cls := "{\n"
	for _, cl := range p.Classes {
		cls += cl.String()
	}
	cls += "}"
	ss := "{\n"
	for _, s := range p.Sources {
		ss += s.String()
	}
	ss += "}"
	cs := "{\n"
	for _, c := range p.Counters {
		cs += c.String()
	}
	cs += "}"
	return fmt.Sprintf("name %s\nclasses %s\nsources %s\ncounters %s\n", p.Name, cls, ss, cs)
}

func (c *Counter) String() string {
	return fmt.Sprintf("type %s, missed %d, covered %d\n", c.Type, c.Missed, c.Covered)
}

func (c *Class) String() string {
	ms := "{\n"
	for _, m := range c.Methods {
		ms += m.String()
	}
	ms += "}"
	cs := "{\n"
	for _, cn := range c.Counters {
		cs += cn.String()
	}
	cs += "}"
	return fmt.Sprintf("name %s\nmethods %s\ncounters %s\n", c.Name, ms, cs)
}

func (s *Source) String() string {
	ls := "{\n"
	for _, l := range s.Lines {
		ls += l.String()
	}
	ls += "}"
	cs := "{\n"
	for _, c := range s.Counters {
		cs += c.String()
	}
	cs += "}"
	return fmt.Sprintf("name %s\nlines %s\ncounters %s\n", s.Name, ls, cs)
}

func (m *Method) String() string {
	cs := "{\n"
	for _, c := range m.Counters {
		cs += c.String()
	}
	cs += "}"
	return fmt.Sprintf("name %s\ndescription %s\nline %d\ncounters %s\n", m.Name, m.Desc, m.Line, cs)
}

func (l *Line) String() string {
	return fmt.Sprintf("NR %d, MI %d, CI %d, MB %d, CB %d\n", l.NR, l.MI, l.CI, l.MB, l.CB)
}
