package db

import (
	"fmt"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"
)

//GetUserById retrieves a user matching the given id from the active database.
func GetUserById(id interface{}) (ret *user.User, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.FindId(id).One(&ret)
	if err != nil {
		err = &DBGetError{"user", err, id}
	}
	return
}

//GetUsers retrieves a user matching the given id from the active database.
func GetUsers(matcher interface{}) (ret []*user.User, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.Find(matcher).Select(bson.M{user.ID: 1}).All(&ret)
	if err != nil {
		err =  &DBGetError{"users", err, matcher}
	}
	return
}

//AddUser adds a new user to the active database.
func AddUser(u *user.User)(err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.Insert(u)
	if err != nil {
		err = &DBAddError{u, err}
	}
	return
}

//AddUsers adds new users to the active database.
func AddUsers(users ...*user.User) (err error){
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.Insert(users)
	if err != nil {
		err = fmt.Errorf("Encountered error %q when adding users %q to db", err, users)
	}
	return
}
