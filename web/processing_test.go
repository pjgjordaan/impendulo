package web

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"

	"testing"
)

func TestSnapshots(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := project.NewSubmission(bson.NewObjectId(), bson.NewObjectId(), "user", project.FILE_MODE, 100000)
	if e := db.Add(db.SUBMISSIONS, s); e != nil {
		t.Error(e)
	}
	specs := []spec{{"Triangle.java", project.SRC, s.Id, 5},
		{"Triangle", project.LAUNCH, s.Id, 3},
		{"Helper.java", project.SRC, s.Id, 4},
		{"UserTests.java", project.TEST, s.Id, 2},
		{"intlola.zip", project.ARCHIVE, s.Id, 1},
	}
	for _, s := range specs {
		testSnapshots(s, t)
	}
}

type spec struct {
	name string
	tipe project.Type
	id   bson.ObjectId
	num  int
}

func testSnapshots(s spec, t *testing.T) {
	for i := 0; i < s.num; i++ {
		if e := db.Add(db.FILES, createFile(s.id, s.tipe, s.name)); e != nil {
			t.Error(e)
		}
	}
	files, e := db.Snapshots(s.id, s.name)
	if e != nil {
		t.Error(e)
	}
	if len(files) != s.num {
		t.Error(fmt.Errorf("Expected %d got %d snapshots.", s.num, len(files)))
	}
}

func createFile(sid bson.ObjectId, t project.Type, n string) *project.File {
	return &project.File{
		Id:    bson.NewObjectId(),
		SubId: sid,
		Name:  n,
		Type:  t,
	}
}
