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
	"github.com/godfried/impendulo/web/calc/stats"
)

type (
	T []map[string]interface{}
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

func User(us []*user.U, ds []*description.D) (T, []*description.D, error) {
	if len(us) == 0 {
		return nil, nil, NoUsersError
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
	return t, UserFields(ds), nil
}

func Project(ps []*project.P, ds []*description.D) (T, []*description.D, error) {
	if len(ps) == 0 {
		return nil, nil, NoProjectsError
	}
	t := New()
	c := stats.NewCalc()
	for _, p := range ps {
		if vs, ns, e := c.Project(p, ds); e == nil {
			m := map[string]interface{}{
				"id": p.Id.Hex(), "name": p.Name, "description": p.Description,
			}
			for i, v := range vs {
				m[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, m)
		}
	}
	return t, ProjectFields(ds), nil
}

func Assignment(as []*project.Assignment, ds []*description.D) (T, []*description.D, error) {
	if len(as) == 0 {
		return nil, nil, NoAssignmentsError
	}
	t := New()
	c := stats.NewCalc()
	for _, a := range as {
		if vs, ns, e := c.Assignment(a, ds); e == nil {
			p := map[string]interface{}{
				"id": a.Id.Hex(), "name": a.Name, "start": a.Start, "end": a.End,
			}
			for i, v := range vs {
				p[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, p)
		}
	}
	return t, AssignmentFields(ds), nil
}

func Submission(ss []*project.Submission, ds []*description.D) (T, []*description.D, error) {
	if len(ss) == 0 {
		return nil, nil, NoSubmissionsError
	}
	t := New()
	c := stats.NewCalc()
	names := make(map[bson.ObjectId]string)
	for _, s := range ss {
		n, ok := names[s.ProjectId]
		if !ok {
			var e error
			if n, e = db.ProjectName(s.ProjectId); e != nil {
				continue
			}
		}
		if vs, ns, e := c.Submission(s, ds); e == nil {
			p := map[string]interface{}{
				"id": s.Id.Hex(), "name": s.User + "'s " + n, "time": s.Time,
			}
			for i, v := range vs {
				p[ds[i].Raw()] = map[string]interface{}{"value": v, "unit": ns[i]}
			}
			t = append(t, p)
		}
	}
	return t, SubmissionFields(ds), nil
}

func SubmissionFields(ds []*description.D) []*description.D {
	return append(description.Ds{{Type: "id"}, {Type: "name"}, {Type: "start date"}, {Type: "start time"}}, ds...)
}

func AssignmentFields(ds []*description.D) []*description.D {
	return append(description.Ds{{Type: "id"}, {Type: "name"}, {Type: "start date"}, {Type: "start time"}, {Type: "end date"}, {Type: "end time"}}, ds...)
}

func ProjectFields(ds []*description.D) []*description.D {
	return append(description.Ds{{Type: "id"}, {Type: "name"}, {Type: "description"}}, ds...)
}

func UserFields(ds []*description.D) []*description.D {
	return append(description.Ds{{Type: "id"}, {Type: "name"}}, ds...)
}

func OverviewFields(ds []*description.D, t string) []*description.D {
	switch t {
	case "user":
		return UserFields(ds)
	case "project":
		return ProjectFields(ds)
	default:
		return ds
	}
}
