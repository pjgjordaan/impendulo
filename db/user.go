package db

import (
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"
)

//GetUserById retrieves a user matching the given id from the active database.
func GetUserById(id interface{}) (ret *user.User, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.FindId(id).One(&ret)
	if err != nil {
		err = &DBGetError{"user", err, id}
	}
	return
}

//GetUsers retrieves users matching the given interface from the active database.
func GetUsers(matcher interface{}) (ret []*user.User, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.Find(matcher).Select(bson.M{user.ID: 1}).All(&ret)
	if err != nil {
		err = &DBGetError{"users", err, matcher}
	}
	return
}

//AddUser adds a new user to the active database.
func AddUser(u *user.User) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.Insert(u)
	if err != nil {
		err = &DBAddError{u.String(), err}
	}
	return
}

//AddUsers adds new users to the active database.
func AddUsers(users ...*user.User) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.Insert(users)
	if err != nil {
		err = fmt.Errorf(
			"Encountered error %q when adding users %q to db",
			err, users)
	}
	return
}

//RemoveUserById removes a user matching
//the given id from the active database.
func RemoveUserById(id interface{}) (err error) {
	subs, err := GetSubmissions(bson.M{project.USER: id},
		bson.M{project.ID: 1})
	if err != nil {
		return
	}
	for _, sub := range subs {
		err = RemoveSubmissionById(sub.Id)
		if err != nil {
			return
		}
	}
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(USERS)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"user", err, id}
	}
	return
}
