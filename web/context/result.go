package context

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"labix.org/v2/mgo/bson"

	"strings"
)

type (
	Result struct {
		Type   string
		Name   string
		FileID bson.ObjectId
	}
)

func NewResult(s string) (*Result, error) {
	r := &Result{}
	if e := r.Set(s); e != nil {
		return nil, e
	}
	return r, nil
}

func (r *Result) Set(s string) error {
	r.Type = ""
	r.Name = ""
	r.FileID = ""
	sp := strings.Split(s, "-")
	if len(sp) > 1 {
		id, e := convert.Id(sp[1])
		if e != nil {
			return e
		}
		r.FileID = id
	}
	sp = strings.Split(sp[0], ":")
	r.Type = sp[0]
	if len(sp) > 1 {
		r.Name = sp[1]
	}
	return nil
}

func (r *Result) Format() string {
	s := r.Type
	if r.Name == "" {
		return s
	}
	s += " \u2192 " + r.Name
	if r.FileID == "" {
		return s
	}
	return s + " \u2192 " + r.Date()
}

func (r *Result) Date() string {
	f, e := db.File(bson.M{db.ID: r.FileID}, bson.M{db.TIME: 1})
	if e != nil {
		return "No File Found"
	}
	return util.Date(f.Time)
}

func (r *Result) Raw() string {
	s := r.Type
	if r.Name == "" {
		return r.Type
	}
	s += ":" + r.Name
	if r.FileID == "" {
		return s
	}
	return s + "-" + r.FileID.Hex()
}

func (r *Result) HasCode() bool {
	return r.Name != ""
}

func (r *Result) Update(sid bson.ObjectId, fname string) error {
	if r.FileID == "" {
		return nil
	}
	id, e := db.FileResultId(sid, fname, r.Type, r.Name)
	if e != nil {
		return r.Set(result.CODE)
	}
	r.FileID = id
	return nil
}

func (r *Result) Check(pid bson.ObjectId) {
	rs := db.ProjectResults(pid)
	cur := r.Raw()
	for _, d := range rs {
		if d == cur || strings.HasPrefix(cur, d) {
			return
		}
	}
	r.Set(result.CODE)
}
