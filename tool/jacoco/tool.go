package jacoco

import (
	"bytes"
	"time"

	"code.google.com/p/go.net/html"

	"encoding/xml"
	"errors"

	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
	"strings"
)

type (
	Tool struct {
		buildPath, resPath string
		test, target       *tool.Target
		testId             bson.ObjectId
	}
)

const (
	NAME = "Jacoco"
)

var (
	ReportError = errors.New("no report created")
	CodeError   = errors.New("no code found")
)

func New(baseDir, srcDir string, test, target *tool.Target, testId bson.ObjectId) (tool.T, error) {
	rd := filepath.Join(baseDir, "target")
	p, e := NewProject("Jacoco Coverage", srcDir, rd, test)
	if e != nil {
		return nil, e
	}
	d, e := xml.Marshal(p)
	if e != nil {
		return nil, e
	}
	bp := filepath.Join(baseDir, "build.xml")
	if e = util.SaveFile(bp, d); e != nil {
		return nil, e
	}
	return &Tool{
		buildPath: bp,
		resPath:   rd,
		test:      test,
		target:    target,
		testId:    testId,
	}, nil
}

func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (result.Tooler, error) {
	if t.target.Executable() != target.Executable() {
		return nil, nil
	}
	if _, e := tool.RunCommand([]string{"ant", "-f", t.buildPath}, nil, 30*time.Second); e != nil {
		return nil, e
	}
	xp := filepath.Join(t.resPath, "report", "report.xml")
	hp := filepath.Join(t.resPath, "report", "html", target.Package, target.FullName()+".html")
	if !util.Exists(xp) || !util.Exists(hp) {
		return nil, ReportError
	}
	hf, e := os.Open(hp)
	if e != nil {
		return nil, e
	}
	d, e := html.Parse(hf)
	if e != nil {
		return nil, e
	}
	n, e := codeNode(d)
	if e != nil {
		return nil, e
	}
	c := new(bytes.Buffer)
	if e = html.Render(c, n); e != nil {
		return nil, e
	}
	xf, e := os.Open(xp)
	if e != nil {
		return nil, e
	}
	return NewResult(fileId, t.testId, t.test.Name, util.ReadBytes(xf), c.Bytes(), target)
}

func codeNode(d *html.Node) (*html.Node, error) {
	var pre *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "pre" {
			for i, a := range n.Attr {
				if a.Key == "class" {
					n.Attr[i].Val = strings.TrimSpace(a.Val + " prettyprint")
					break
				}
			}
			pre = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(d)
	if pre == nil {
		return nil, CodeError
	}
	return pre, nil
}

func (t *Tool) Name() string {
	return NAME + ":" + t.test.Name
}

func (t *Tool) Lang() project.Language {
	return project.JAVA
}
