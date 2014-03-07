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
	}
)

const (
	NAME = "Jacoco"
)

func New(baseDir, srcDir string, test *tool.Target) (t tool.Tool, err error) {
	resDir := filepath.Join(baseDir, "target")
	p, err := NewProject("Jacoco Coverage", srcDir, resDir, test)
	if err != nil {
		return
	}
	data, err := xml.Marshal(p)
	if err != nil {
		return
	}
	buildPath := filepath.Join(baseDir, "build.xml")
	err = util.SaveFile(buildPath, data)
	if err != nil {
		return
	}
	t = &Tool{
		buildPath: buildPath,
		resPath:   resDir,
		test:      test,
	}
	return
}

func (this *Tool) Run(fileId bson.ObjectId, target *tool.Target) (res tool.ToolResult, err error) {
	execRes := tool.RunCommand([]string{"ant", "-f", this.buildPath}, nil)
	if execRes.Err != nil {
		err = execRes.Err
		return
	}
	xmlPath := filepath.Join(this.resPath, "report", "report.xml")
	htmlPath := filepath.Join(this.resPath, "report", "html", target.Package, target.FullName()+".html")
	if !util.Exists(xmlPath) || !util.Exists(htmlPath) {
		err = errors.New("no report created")
		util.Log(err, string(execRes.StdErr))
		return
	}
	htmlFile, err := os.Open(htmlPath)
	if err != nil {
		return
	}
	d, err := html.Parse(htmlFile)
	if err != nil {
		return
	}
	n, err := codeNode(d)
	if err != nil {
		return
	}
	code := new(bytes.Buffer)
	err = html.Render(code, n)
	if err != nil {
		return
	}
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		return
	}
	data := util.ReadBytes(xmlFile)
	res, err = NewResult(fileId, this.test.Name, data, code.Bytes())
	return

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

func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) Lang() tool.Language {
	return tool.JAVA
}
