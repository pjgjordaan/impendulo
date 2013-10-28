package webserver

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
	s := project.NewSubmission(bson.NewObjectId(), "user", project.FILE_MODE, 100000)
	err := db.Add(db.SUBMISSIONS, s)
	if err != nil {
		t.Error(err)
	}
	src, launch, other := 5, 3, 7
	for i := 0; i < src; i++ {
		tsrc := createFile(s.Id, project.SRC, "Triangle.java")
		err = db.Add(db.FILES, tsrc)
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < launch; i++ {
		tlaunch := createFile(s.Id, project.LAUNCH, "Triangle.java")
		err = db.Add(db.FILES, tlaunch)
		if err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < other; i++ {
		othersrc := createFile(s.Id, project.SRC, "Helper.java")
		err = db.Add(db.FILES, othersrc)
		if err != nil {
			t.Error(err)
		}
	}
	snaps, err := Snapshots(s.Id, "Triangle.java")
	if err != nil {
		t.Error(err)
	}
	if len(snaps) != src {
		t.Error(fmt.Errorf("Expected %d got %d snapshots.", src, len(snaps)))
	}
}

func createFile(subId bson.ObjectId, tipe project.Type, name string) *project.File {
	return &project.File{
		Id:    bson.NewObjectId(),
		SubId: subId,
		Name:  name,
		Type:  tipe,
	}
}
