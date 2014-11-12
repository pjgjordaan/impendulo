//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package tables

import (
	"errors"

	"labix.org/v2/mgo/bson"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/result/description"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util/milliseconds"
	"github.com/godfried/impendulo/web/calc/stats"
)

type (
	T     []map[string]interface{}
	pname map[bson.ObjectId]string
)

var (
	NoFilesError       = errors.New("no files to load table for")
	NoSubmissionsError = errors.New("no submissions to create table for")
	NoAssignmentsError = errors.New("no assignments to create table for")
	NoProjectsError    = errors.New("no projects to create table for")
	NoUsersError       = errors.New("no users to create table for")
)

func New() T {
	return make(T, 0, 1000)
}

func User(us []*user.U, ds description.Ds) (T, error) {
	if len(us) == 0 {
		return nil, NoUsersError
	}
	t := New()
	c := stats.NewCalc()
	for _, u := range us {
		if vs, ns, e := c.User(u, ds); e == nil {
			m := map[string]interface{}{
				"id": u.Name, "name": u.Name,
			}
			for i, v := range vs {
				m[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, m)
		}
	}
	return t, nil
}

func Project(ps []*project.P, ds description.Ds) (T, error) {
	if len(ps) == 0 {
		return nil, NoProjectsError
	}
	t := New()
	c := stats.NewCalc()
	for _, p := range ps {
		if vs, ns, e := c.Project(p, ds); e == nil {
			m := map[string]interface{}{
				"id": p.Id.Hex(), "name": p.Name, "description": p.Description, "author": p.User, "language": p.Lang,
			}
			for i, v := range vs {
				m[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, m)
		}
	}
	return t, nil
}

func Assignment(as []*project.Assignment, ds description.Ds) (T, error) {
	if len(as) == 0 {
		return nil, NoAssignmentsError
	}
	t := New()
	c := stats.NewCalc()
	pns := make(pname)
	for _, a := range as {
		pn := pns.get(a.ProjectId)
		if pn == "" {
			continue
		}
		if vs, ns, e := c.Assignment(a, ds); e == nil {
			p := map[string]interface{}{
				"id": a.Id.Hex(), "name": a.Name, "project": pn, "author": a.User, "start date": milliseconds.DateString(a.Start),
				"start time": milliseconds.TimeString(a.Start), "end date": milliseconds.DateString(a.End), "end time": milliseconds.TimeString(a.End),
			}
			for i, v := range vs {
				p[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, p)
		}
	}
	return t, nil
}

func Submission(ss []*project.Submission, ds description.Ds) (T, error) {
	if len(ss) == 0 {
		return nil, NoSubmissionsError
	}
	t := New()
	c := stats.NewCalc()
	pns := make(pname)
	for _, s := range ss {
		pn := pns.get(s.ProjectId)
		if pn == "" {
			continue
		}
		if vs, ns, e := c.Submission(s, ds); e == nil {
			p := map[string]interface{}{
				"id": s.Id.Hex(), "user": s.User, "project": pn, "start date": milliseconds.DateString(s.Time), "start time": milliseconds.TimeString(s.Time),
			}
			for i, v := range vs {
				p[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, p)
		}
	}
	return t, nil
}

func File(fs []*project.File, ds description.Ds) (T, error) {
	if len(fs) == 0 {
		return nil, NoFilesError
	}
	t := New()
	c := stats.NewCalc()
	s, e := db.Submission(bson.M{db.ID: fs[0].SubId}, nil)
	if e != nil {
		return nil, e
	}
	pn, e := db.ProjectName(s.ProjectId)
	if e != nil {
		return nil, e
	}
	for _, f := range fs {
		if vs, ns, e := c.File(f, ds); e == nil {
			p := map[string]interface{}{
				"id": f.Name, "name": f.Name, "package": f.Package, "type": f.Type.Title(), "user": s.User, "project": pn, "date": milliseconds.DateString(f.Time), "time": milliseconds.TimeString(f.Time),
			}
			for i, v := range vs {
				p[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, p)
		}
	}
	return t, nil
}

func FileFields() description.Ds {
	return description.Ds{{Type: "id"}, {Type: "name"}, {Type: "package"}, {Type: "type"}, {Type: "user"}, {Type: "project"}, {Type: "date"}, {Type: "time"}}
}

func SubmissionFields() description.Ds {
	return description.Ds{{Type: "id"}, {Type: "user"}, {Type: "project"}, {Type: "start date"}, {Type: "start time"}}
}

func AssignmentFields() description.Ds {
	return description.Ds{{Type: "id"}, {Type: "name"}, {Type: "project"}, {Type: "author"}, {Type: "start date"}, {Type: "start time"}, {Type: "end date"}, {Type: "end time"}}
}

func ProjectFields() description.Ds {
	return description.Ds{{Type: "id"}, {Type: "name"}, {Type: "description"}, {Type: "author"}, {Type: "language"}}
}

func UserFields() description.Ds {
	return description.Ds{{Type: "id"}, {Type: "name"}}
}

func OverviewFields(t string) []*description.D {
	switch t {
	case "user":
		return UserFields()
	case "project":
		return ProjectFields()
	default:
		return description.Ds{}
	}
}

func (p pname) get(id bson.ObjectId) string {
	n, ok := p[id]
	if ok {
		return n
	}
	p[id], _ = db.ProjectName(id)
	return p[id]
}
