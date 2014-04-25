package jacoco

import (
	"bytes"

	"code.google.com/p/go.net/html"

	"encoding/xml"
	"errors"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
	"strings"
)

type (
	Tool struct {
		buildPath, resPath string
		test               *tool.Target
		testId             bson.ObjectId
	}
)

const (
	NAME = "Jacoco"
)

func New(baseDir, srcDir string, test *tool.Target, testId bson.ObjectId) (tool.Tool, error) {
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
		testId:    testId,
	}, nil
}

func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (tool.ToolResult, error) {
	if _, e := tool.RunCommand([]string{"ant", "-f", t.buildPath}, nil); e != nil {
		return nil, e
	}
	xp := filepath.Join(t.resPath, "report", "report.xml")
	hp := filepath.Join(t.resPath, "report", "html", target.Package, target.FullName()+".html")
	if !util.Exists(xp) || !util.Exists(hp) {
		return nil, errors.New("no report created")
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
		return nil, errors.New("no code found")
	}
	return pre, nil
}

func (t *Tool) Name() string {
	return NAME + ":" + t.test.Name
}

func (t *Tool) Lang() tool.Language {
	return tool.JAVA
}
