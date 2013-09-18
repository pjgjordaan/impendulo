package db

import (
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"
)

//User retrieves a user matching the given id from the active database.
func User(id interface{}) (ret *user.User, err error) {
	session, err := Session()
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

//Users retrieves users matching the given interface from the active database.
func Users(matcher interface{}, sort ...string) (ret []*user.User, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(USERS)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(bson.M{user.ID: 1}).All(&ret)
	if err != nil {
		err = &DBGetError{"users", err, matcher}
	}
	return
}

//AddUsers adds new users to the active database.
func AddUsers(users ...*user.User) (err error) {
	session, err := Session()
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
	subs, err := Submissions(bson.M{project.USER: id},
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
	err = RemoveById(USERS, id)
	return
}
