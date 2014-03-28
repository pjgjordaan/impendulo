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
func User(id string) (*user.User, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	c := s.DB("").C(USERS)
	var u *user.User
	if e = c.FindId(id).One(&u); e != nil {
		return nil, &DBGetError{"user", e, id}
	}
	return u, nil
}

//Users retrieves users matching the given interface from the active database.
func Users(matcher interface{}, sort ...string) ([]*user.User, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	c := s.DB("").C(USERS)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	var u []*user.User
	if e = q.Select(bson.M{user.ID: 1}).All(&u); e != nil {
		return nil, &DBGetError{"users", e, matcher}
	}
	return u, nil
}

func UserNames(matcher interface{}) ([]string, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	c := s.DB("").C(USERS)
	var n []string
	if e = c.Find(matcher).Sort(NAME).Distinct(NAME, &n); e != nil {
		return nil, &DBGetError{"users", e, matcher}
	}
	return n, nil
}

//AddUsers adds new users to the active database.
func AddUsers(users ...*user.User) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	c := s.DB("").C(USERS)
	if e = c.Insert(users); e != nil {
		return fmt.Errorf("error %q adding users %q to db", e, users)
	}
	return nil
}

//RemoveUserById removes a user matching
//the given id from the active database.
func RemoveUserById(id string) error {
	ss, e := Submissions(bson.M{USER: id}, bson.M{ID: 1})
	if e != nil {
		return e
	}
	for _, s := range ss {
		if e = RemoveSubmissionById(s.Id); e != nil {
			return e
		}
	}
	return RemoveById(USERS, id)
}

func RenameUser(o, n string) error {
	u, e := User(o)
	if e != nil {
		return e
	}
	u.Name = n
	if e = Add(USERS, u); e != nil {
		return e
	}
	m := bson.M{USER: o}
	c := bson.M{SET: bson.M{USER: n}}
	if e = UpdateAll(PROJECTS, m, c); e != nil {
		return e
	}
	if e = UpdateAll(SUBMISSIONS, m, c); e != nil {
		return e
	}
	return RemoveUserById(o)
}
