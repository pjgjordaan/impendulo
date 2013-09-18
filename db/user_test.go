package db

import (
	"github.com/godfried/impendulo/user"
	"reflect"
	"testing"
)

func TestUser(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	u := user.New("uname", "pword")
	err = Add(USERS, u)
	if err != nil {
		t.Error(err)
	}
	found, err := User("uname")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(u, found) {
		t.Error("Users not equivalent", u, found)
	}
}
