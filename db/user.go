//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
