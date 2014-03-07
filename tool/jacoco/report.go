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
		Id       bson.ObjectId
		Name     string     `xml:"name,attr"`
		Packages []*Package `xml:"package"`
		Counters []*Counter `xml:"counter"`
		HTML     []byte
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
func NewReport(id bson.ObjectId, xmlData, htmlData []byte) (res *Report, err error) {
	if err = xml.Unmarshal(xmlData, &res); err != nil {
		if res == nil {
			err = tool.NewXMLError(err, "jacoco/jacocoResult.go")
			return
		} else {
			err = nil
		}
	}
	res.HTML = htmlData
	res.Id = id
	return
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
