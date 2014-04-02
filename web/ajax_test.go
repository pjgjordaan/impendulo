package web

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"

	"testing"
)

type (
	triple struct {
		p    float64
		i, n int
	}
)

func TestDetermineName(t *testing.T) {
	tests := []triple{{0, 99, 100}, {0.9, 0, 100}, {0.5, 50, 100}, {0.3, 70, 100}, {0.001, 12, 100}}
	for _, c := range tests {
		if e := testDetermineName(c.p, c.i, c.n); e != nil {
			t.Error(e)
		}
	}

}

func testDetermineName(p float64, i, n int) error {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := project.NewSubmission(bson.NewObjectId(), "user", project.FILE_MODE, 10000)
	if e := db.Add(db.SUBMISSIONS, s); e != nil {
		return e
	}
	cf := project.File{
		Id:    bson.NewObjectId(),
		SubId: s.Id,
		Name:  "Testee",
		Time:  s.Time + 25,
	}
	if e := db.Add(db.FILES, cf); e != nil {
		return e
	}
	r := "Tool-" + cf.Id.Hex()
	for j := 0; j < n; j++ {
		f := project.File{
			Id:    bson.NewObjectId(),
			SubId: s.Id,
			Name:  "Test",
			Time:  s.Time + int64((i+12)*23),
		}
		if j == i {
			f.Results = bson.M{r: ""}
		}
		if e := db.Add(db.FILES, f); e != nil {
			return e
		}
	}
	rn, e := determineName(s.Id, "Test", "Testee", "Tool", p)
	if e != nil {
		return e
	}
	if rn != r {
		return fmt.Errorf("incorrect name %s expected %s", s, r)
	}
	return nil
}
