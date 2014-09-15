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

//Package db provides an interface to mgo which allows us to store and retrieve Impendulo
//data from mongodb.
package db

import (
	"fmt"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	sessionChan chan *mgo.Session
	requestChan chan bool
)

//Setup creates a mongodb session.
//This must be called before using any other db functions.
func Setup(c string) error {
	s, e := mgo.Dial(c)
	if e != nil {
		return e
	}
	sessionChan = make(chan *mgo.Session)
	requestChan = make(chan bool)
	go serveSession(s)
	return nil
}

//serveSession manages the active session.
func serveSession(s *mgo.Session) {
	for {
		r, ok := <-requestChan
		if !ok || !r {
			break
		}
		if s == nil {
			sessionChan <- nil
		} else {
			sessionChan <- s.Clone()
		}
	}
	if s != nil {
		s.Close()
	}
	close(requestChan)
	close(sessionChan)
}

//Session retrieves the current active session.
func Session() (*mgo.Session, error) {
	requestChan <- true
	s, ok := <-sessionChan
	if s == nil || !ok {
		return nil, fmt.Errorf("could not retrieve session")
	}
	return s, nil
}

//Close shuts down the current session.
func Close() {
	requestChan <- false
}

//DeleteDB removes a db.
func DeleteDB(db string) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	return s.DB(db).DropDatabase()
}

//CopyDB is used to copy the contents of one database to a new location.
func CopyDB(f, t string) error {
	if e := DeleteDB(f); e != nil {
		return e
	}
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	return s.Run(bson.D{{"copydb", "1"}, {"fromdb", f}, {"todb", t}}, nil)
}

//CloneCollection
func CloneCollection(o, c string) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	return s.DB("").Run(bson.D{{"cloneCollection", c}, {"from", o}}, nil)
}

//CloneData
func CloneData(o string) error {
	cs := []string{USERS, PROJECTS, SUBMISSIONS, FILES, TESTS, JPF, PMD}
	for _, c := range cs {
		if e := CloneCollection(o, c); e != nil {
			return e
		}
	}
	return nil
}

//Add adds a document to the specified collection.
func Add(n string, i interface{}) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	if e = s.DB("").C(n).Insert(i); e != nil {
		return &AddError{n, e}
	}
	return nil
}

//RemoveById removes a document matching the given id in collection n from the active database.
func RemoveById(n string, id interface{}) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	if e = s.DB("").C(n).RemoveId(id); e != nil {
		return &RemoveError{n, e, id}
	}
	return nil
}

//Update updates documents from the collection n matching m with the changes specified by c.
func Update(n string, m, c interface{}) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	if e = s.DB("").C(n).Update(m, c); e != nil {
		return fmt.Errorf("error %q: updating %q matching %q to %q", e, n, m, c)
	}
	return nil
}

func UpdateAll(n string, m, c interface{}) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	if _, e = s.DB("").C(n).UpdateAll(m, c); e != nil {
		return fmt.Errorf("error %q: updating %q matching %q to %q", e, n, m, c)
	}
	return nil
}

//Contains checks whether the collection n contains any items matching m.
func Contains(n string, m interface{}) bool {
	c, e := Count(n, m)
	return e == nil && c > 0
}

//Count calculates the amount of items in the collection col which match matcher.
func Count(n string, m interface{}) (int, error) {
	s, e := Session()
	if e != nil {
		return -1, e
	}
	defer s.Close()
	count, e := s.DB("").C(n).Find(m).Count()
	if e != nil {
		return -1, &GetError{n + " count", e, m}
	}
	return count, nil
}

func Collections(db string) ([]string, error) {
	s, e := mgo.Dial(ADDRESS + db)
	if e != nil {
		return nil, e
	}
	defer s.Close()
	return s.DB("").CollectionNames()
}

func Databases() ([]string, error) {
	s, e := mgo.Dial(ADDRESS)
	if e != nil {
		return nil, e
	}
	defer s.Close()
	return s.DatabaseNames()
}

func IDs(c string, m bson.M) ([]bson.ObjectId, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var ids []bson.ObjectId
	if e := s.DB("").C(c).Find(m).Distinct(ID, &ids); e != nil {
		return nil, &GetError{"ids", e, m}
	}
	return ids, nil
}
