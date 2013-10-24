//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package db

import (
	"fmt"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"
)

//User retrieves a user matching the given id from the active database.
func User(id string) (ret *user.User, err error) {
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
func RemoveUserById(id string) (err error) {
	subs, err := Submissions(bson.M{USER: id},
		bson.M{ID: 1})
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

func RenameUser(oldName, newName string) (err error) {
	u, err := User(oldName)
	if err != nil {
		return
	}
	u.Name = newName
	err = Add(USERS, u)
	if err != nil {
		return
	}
	change := bson.M{SET: bson.M{USER: newName}}
	matcher := bson.M{USER: oldName}
	err = UpdateAll(PROJECTS, matcher, change)
	if err != nil {
		return
	}
	err = UpdateAll(SUBMISSIONS, matcher, change)
	if err != nil {
		return
	}
	err = RemoveUserById(oldName)
	return
}
