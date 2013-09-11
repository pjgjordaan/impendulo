package checkstyle

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
	"math"
	"strings"
)

const NAME = "Checkstyle"

func init() {
	gob.Register(new(Report))
}

type Result struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	Data   *Report       "data"
	GridFS bool          "gridfs"
}

func (this *Result) SetData(data interface{}) {
	if data == nil {
		this.Data = nil
	} else {
		this.Data = data.(*Report)
	}
}

func (this *Result) OnGridFS() bool {
	return this.GridFS
}

func (this *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; Name: %s; \nData: %s\n",
		this.Id, this.FileId, this.Name, this.Data.String())
}

func (this *Result) GetName() string {
	return this.Name
}

func (this *Result) GetSummary() *tool.Summary {
	body := fmt.Sprintf("Errors: %d",
		this.Data.Errors)
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Result) GetData() interface{} {
	return this.Data
}

func (this *Result) Template(current bool) string {
	if current {
		return "checkstyleCurrent"
	} else {
		return "checkstyleNext"
	}
}

func (this *Result) Success() bool {
	return true
}

func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart("Checkstyle Errors")
	}
	y := float64(this.Data.Errors)
	tool.AddCoords(graphData[0], x, y)
	return math.Max(max, y)
}

func NewResult(fileId bson.ObjectId, data []byte) (res *Result, err error) {
	gridFS := len(data) > tool.MAX_SIZE
	res = &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		GridFS: gridFS,
	}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *Report, err error) {
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

type Report struct {
	Id      bson.ObjectId
	Version string `xml:"version,attr"`
	Errors  int
	Files   []*File `xml:"file"`
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

type File struct {
	Name   string   `xml:"name,attr"`
	Errors []*Error `xml:"error"`
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

type Problem struct {
	*Error
	Lines []int
}

type Error struct {
	Line     int           `xml:"line,attr"`
	Column   int           `xml:"column,attr"`
	Severity string        `xml:"severity,attr"`
	Message  template.HTML `xml:"message,attr"`
	Source   string        `xml:"source,attr"`
}

func (this *Error) String() string {
	return fmt.Sprintf("Line: %d; Column: %d; Severity: %s; Message: %q; Source: %s\n",
		this.Line, this.Column, this.Severity, this.Message, this.Source)
}
