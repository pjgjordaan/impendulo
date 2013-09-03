package javac

import (
	"bytes"
	"errors"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"math"
	"strconv"
)

const NAME = "Javac"

type Result struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	Data   []byte        "data"
}

func (this *Result) GetName() string {
	return this.Name
}

func (this *Result) GetId() bson.ObjectId {
	return this.Id
}

func (this *Result) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *Result) GetSummary() *tool.Summary {
	var body string
	if this.Success() {
		body = "Compiled successfully."
	} else {
		body = "No compile."
	}
	return &tool.Summary{
		Name: this.GetName(),
		Body: body,
	}
}

func (this *Result) GetData() interface{} {
	return this
}

func (this *Result) Template(current bool) string {
	if current {
		return "javacCurrent"
	} else {
		return "javacNext"
	}
}

var (
	compSuccess  = []byte("Compiled successfully")
	compWarning  = []byte("warning")
	compWarnings = []byte("warnings")
	compError    = []byte("error")
	compErrors   = []byte("errors")
)

func (this *Result) Success() bool {
	return bytes.Equal(this.Data, compSuccess)
}

func (this *Result) Warnings() bool {
	return bytes.HasSuffix(this.Data, compWarning) ||
		bytes.HasSuffix(this.Data, compWarnings)
}

func (this *Result) Errors() bool {
	return bytes.HasSuffix(this.Data, compError) ||
		bytes.HasSuffix(this.Data, compErrors)
}

func (this *Result) Count() (n int, err error) {
	if this.Success() {
		err = errors.New("No count for successfull compile.")
		return
	}
	split := bytes.Split(this.Data, []byte("\n"))
	if len(split) < 1 {
		err = errors.New("Can't find count line in message.")
		return
	}
	split = bytes.Split(bytes.TrimSpace(split[len(split)-1]), []byte(" "))
	if len(split) < 1 {
		err = errors.New("Can't find count in last line.")
		return
	}
	n, err = strconv.Atoi(string(split[0]))
	return
}

func (this *Result) ResultHeader() (header string) {
	if this.Success() {
		header = string(this.Data)
		return
	} else {
		count, err := this.Count()
		if err != nil {
			return "Could not retrieve compilation result."
		}
		header = strconv.Itoa(count) + " "
		if this.Warnings() {
			header += "Warning"
		} else if this.Errors() {
			header += "Error"
		}
		if count > 1 {
			header += "s"
		}
	}
	return
}

func (this *Result) Result() string {
	return string(this.Data)
}

func (this *Result) AddGraphData(max, x float64, graphData []map[string]interface{}) float64 {
	if graphData[0] == nil {
		graphData[0] = tool.CreateChart(this.GetName() + " Errors")
		graphData[1] = tool.CreateChart(this.GetName() + " Warnings")
	}
	yE, yW := 0.0, 0.0
	if this.Errors() {
		n, err := this.Count()
		if err == nil {
			yE = float64(n)
		}
	} else if this.Warnings() {
		n, err := this.Count()
		if err == nil {
			yW = float64(n)
		}
	}
	tool.AddCoords(graphData[0], x, yE)
	tool.AddCoords(graphData[1], x, yW)
	return math.Max(max, math.Max(yE, yW))
}

func NewResult(fileId bson.ObjectId, data []byte) *Result {
	return &Result{
		Id:     bson.NewObjectId(),
		FileId: fileId,
		Name:   NAME,
		Data:   bytes.TrimSpace(data),
	}
}
