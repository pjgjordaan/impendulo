package db

import (
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
	"strconv"
	"testing"
)

func TestSetup(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
}

func TestCount(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	num := 100
	n, err := Count(PROJECTS, bson.M{})
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("Invalid count %q, should be %q", n, 0)
	}
	for i := 0; i < num; i++ {
		var s int = i / 10
		err = Add(PROJECTS, project.NewProject("name"+strconv.Itoa(s), "user", "lang", []byte{}))
		if err != nil {
			t.Error(err)
		}
	}
	n, err = Count(PROJECTS, bson.M{})
	if err != nil {
		t.Error(err)
	}
	if n != num {
		t.Errorf("Invalid count %q, should be %q", n, num)
	}
	n, err = Count(PROJECTS, bson.M{"name": "name0"})
	if err != nil {
		t.Error(err)
	}
	if n != 10 {
		t.Errorf("Invalid count %q, should be %q", n, 10)
	}

}
