package description

import (
	"encoding/json"
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/code"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/util/milliseconds"
	"labix.org/v2/mgo/bson"

	"strings"
)

type (
	D struct {
		Type   string
		Name   string
		FileID bson.ObjectId
		Metric string
	}
	Ds []*D
)

const (
	M_SEP = "~"
	F_SEP = "-"
	N_SEP = ":"
)

func New(s string) (*D, error) {
	d := &D{}
	if e := d.Set(s); e != nil {
		return nil, e
	}
	return d, nil
}

func (d *D) String() string {
	return fmt.Sprintf("Type:%s Name:%s FileID:%s Metric:%s", d.Type, d.Name, d.FileID, d.Metric)
}

func (d *D) Set(s string) error {
	d.Type = ""
	d.Name = ""
	d.FileID = ""
	d.Metric = ""
	sp := strings.Split(s, M_SEP)
	if len(sp) > 1 {
		d.Metric = sp[1]
	}
	sp = strings.Split(sp[0], F_SEP)
	if len(sp) > 1 {
		id, e := convert.Id(sp[1])
		if e != nil {
			return e
		}
		d.FileID = id
	}
	sp = strings.Split(sp[0], N_SEP)
	d.Type = sp[0]
	if len(sp) > 1 {
		d.Name = sp[1]
	}
	return nil
}

func (d *D) Format() string {
	s := d.Type
	if d.Name != "" {
		s += " \u2192 " + d.Name
	}
	if d.FileID != "" {
		s += " \u2192 " + d.Date()
	}
	if d.Metric != "" {
		s += " \u2192 " + d.Metric
	}
	return s
}

func (d *D) Date() string {
	f, e := db.File(bson.M{db.ID: d.FileID}, db.FILE_SELECTOR)
	if e != nil {
		return "No File Found"
	}
	return milliseconds.DateTimeString(f.Time)
}

func (d *D) Key() string {
	s := d.Type
	if d.Name != "" {
		s += N_SEP + d.Name
	}
	if d.FileID != "" {
		s += F_SEP + d.FileID.Hex()
	}
	return s
}

func (d *D) Raw() string {
	s := d.Key()
	if d.Metric != "" {
		s += M_SEP + d.Metric
	}
	return s
}

func (d *D) HasCode() bool {
	return d.Name != ""
}

func (d *D) Update(sid bson.ObjectId, fname string) error {
	if d.FileID == "" {
		return nil
	}
	id, e := db.FileResultId(sid, fname, d.Type, d.Name)
	if e != nil {
		return d.Set(code.NAME)
	}
	d.FileID = id
	return nil
}

func (d *D) Check(pid bson.ObjectId) {
	rs := db.ProjectResults(pid)
	cur := d.Key()
	for _, d := range rs {
		if d == cur || strings.HasPrefix(cur, d) {
			return
		}
	}
	d.Set(code.NAME)
}

func (d *D) Charter(f *project.File) (result.Charter, error) {
	if _, e := convert.Id(f.Results[d.Key()]); e != nil {
		return d.charter(f)
	}
	return db.Charter(bson.M{db.ID: f.Results[d.Key()]}, nil)
}

func (d *D) charter(f *project.File) (result.Charter, error) {
	switch d.Type {
	case code.NAME:
		fd, e := db.File(bson.M{db.ID: f.Id}, nil)
		if e != nil {
			return nil, e
		}
		return code.New(fd.Id, project.JAVA, fd.Data), nil
	default:
		return nil, fmt.Errorf("not a charter %s", d.Type)
	}
}

func (ds Ds) Len() int {
	return len(ds)
}

func (ds Ds) Swap(i, j int) {
	ds[i], ds[j] = ds[j], ds[i]
}

func (ds Ds) Less(i, j int) bool {
	return strings.ToLower(ds[i].Raw()) <= strings.ToLower(ds[j].Raw())
}

func (d *D) MarshalJSON() ([]byte, error) {
	return []byte(`{"id":"` + d.Raw() + `", "name":"` + d.Format() + `"}`), nil
}

func (d *D) UnmarshalJSON(b []byte) error {
	type j struct {
		Id string `json:"id"`
	}
	var id *j
	if e := json.Unmarshal(b, &id); e != nil {
		return e
	}
	return d.Set(id.Id)
}
