package db

import (
	"errors"

	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/web/calc"
	"labix.org/v2/mgo/bson"
)

func Table(did string, l calc.Level) (*calc.T, error) {
	return table(bson.M{DATAID: did, LEVEL: l, TYPE: calc.TABLE}, nil)
}

func table(m, sl interface{}) (*calc.T, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var t *calc.T
	if e = s.DB("").C(CALC).Find(m).Select(sl).One(&t); e != nil {
		return nil, &GetError{"calc", e, m}
	}
	return t, nil
}

func Chart(did string, l calc.Level, x, y string) (*calc.C, error) {
	return chart(bson.M{DATAID: did, TYPE: calc.CHART, LEVEL: l, X: x, Y: y}, nil)
}

func chart(m, sl interface{}) (*calc.C, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var c *calc.C
	if e = s.DB("").C(CALC).Find(m).Select(sl).One(&c); e != nil {
		return nil, &GetError{"calc", e, m}
	}
	return c, nil
}

func AddTable(data []map[string]interface{}, l calc.Level, did string) error {
	m := bson.M{DATAID: did, TYPE: calc.TABLE, LEVEL: l}
	if Contains(CALC, m) {
		return Update(CALC, m, bson.M{DATA: data})
	}
	return Add(CALC, &calc.T{Id: bson.NewObjectId(), Type: calc.TABLE, Level: l, DataId: did, Data: data})
}

func AddChart(data []map[string]interface{}, info map[string]interface{}, l calc.Level, did, x, y string) error {
	m := bson.M{DATAID: did, TYPE: calc.CHART, LEVEL: l, X: x, Y: y}
	if Contains(CALC, m) {
		return Update(CALC, m, bson.M{DATA: data, INFO: info})
	}
	return Add(CALC, &calc.C{Id: bson.NewObjectId(), Info: info, Type: calc.CHART, Level: l, DataId: did, Data: data, X: x, Y: y})
}

func RemoveCalcsU(c string, m bson.M) error {
	switch c {
	case USERS:
		us, e := Users(m)
		if e != nil {
			return e
		}
		for _, u := range us {
			removeUserCalc(u.Name)
		}
	case PROJECTS:
		ps, e := Projects(m, bson.M{ID: 1})
		if e != nil {
			return e
		}
		for _, p := range ps {
			removeProjectCalc(p.Id)
		}
	case ASSIGNMENTS:
		as, e := Assignments(m, nil)
		if e != nil {
			return e
		}
		for _, a := range as {
			removeAssignmentCalc(a)
		}
	case SUBMISSIONS:
		ss, e := Submissions(m, nil)
		if e != nil {
			return e
		}
		for _, s := range ss {
			removeSubmissionCalc(s)
		}
	case FILES:
		fs, e := Files(m, bson.M{DATA: 0, RESULTS: 0}, 0)
		if e != nil {
			return e
		}
		for _, f := range fs {
			removeFileCalc(f)
		}
	case RESULTS:
		r, e := Tooler(m, bson.M{ID: 1, FILEID: 1})
		if e != nil {
			return e
		}
		return removeResultCalc(r)
	}
	return nil
}

func RemoveCalcs(c string, i interface{}) error {
	switch c {
	case PROJECTS:
		p, ok := i.(*project.P)
		if !ok {
			return errors.New("could not cast project")
		}
		removeProjectCalc(p.Id)
	case ASSIGNMENTS:
		a, ok := i.(*project.Assignment)
		if !ok {
			return errors.New("could not cast assignment")
		}
		removeAssignmentCalc(a)
	case SUBMISSIONS:
		s, ok := i.(*project.Submission)
		if !ok {
			return errors.New("could not cast submission")
		}
		return removeSubmissionCalc(s)
	case FILES:
		f, ok := i.(*project.File)
		if !ok {
			return errors.New("could not cast file")
		}
		return removeFileCalc(f)
	case USERS:
		u, ok := i.(*user.U)
		if !ok {
			return errors.New("could not cast user")
		}
		removeUserCalc(u.Name)
	case RESULTS:
		r, ok := i.(result.Tooler)
		if !ok {
			return errors.New("could not cast result")
		}
		return removeResultCalc(r)
	}
	return nil
}

func removeResultCalc(r result.Tooler) error {
	Remove(CALC, bson.M{DATAID: r.GetId()})
	f, e := File(bson.M{ID: r.GetFileId()}, bson.M{DATA: 0, RESULTS: 0})
	if e != nil {
		return e
	}
	return removeFileCalc(f)
}

func removeFileCalc(f *project.File) error {
	Remove(CALC, bson.M{DATAID: f.Id})
	s, e := Submission(bson.M{ID: f.SubId}, nil)
	if e != nil {
		return e
	}
	return removeSubmissionCalc(s)
}

func removeSubmissionCalc(s *project.Submission) error {
	Remove(CALC, bson.M{DATAID: s.Id})
	removeUserCalc(s.User)
	a, e := Assignment(bson.M{ID: s.AssignmentId}, nil)
	if e != nil {
		return e
	}
	removeAssignmentCalc(a)
	return nil
}

func removeAssignmentCalc(a *project.Assignment) {
	Remove(CALC, bson.M{DATAID: a.Id})
	removeProjectCalc(a.ProjectId)
}

func removeProjectCalc(id bson.ObjectId) {
	Remove(CALC, bson.M{DATAID: id})
	Remove(CALC, bson.M{DATAID: "project"})
}

func removeUserCalc(id string) {
	Remove(CALC, bson.M{DATAID: id})
	Remove(CALC, bson.M{DATAID: "user"})
}
