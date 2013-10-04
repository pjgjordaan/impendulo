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

const (
	//Mongodb collection name.
	USERS       = "users"
	SUBMISSIONS = "submissions"
	FILES       = "files"
	RESULTS     = "results"
	TESTS       = "tests"
	PROJECTS    = "projects"
	JPF         = "jpf"
	PMD         = "pmd"
	//Mongodb command
	SET = "$set"
	//Mongodb connection and db names
	ADDRESS      = "mongodb://localhost/"
	DEFAULT_DB   = "impendulo"
	DEBUG_DB     = "impendulo_debug"
	TEST_DB      = "impendulo_test"
	BACKUP_DB    = "impendulo_backup"
	DEFAULT_CONN = ADDRESS + DEFAULT_DB
	DEBUG_CONN   = "mongodb://localhost/impendulo_debug"
	TEST_CONN    = "mongodb://localhost/impendulo_test"
)

var (
	sessionChan chan *mgo.Session
	requestChan chan bool
)

//Setup creates a mongodb session.
//This must be called before using any other db functions.
func Setup(conn string) error {
	activeSession, err := mgo.Dial(conn)
	if err != nil {
		return err
	}
	sessionChan = make(chan *mgo.Session)
	requestChan = make(chan bool)
	go serveSession(activeSession)
	return nil
}

//serveSession manages the active session.
func serveSession(activeSession *mgo.Session) {
	for {
		req, ok := <-requestChan
		if !ok || !req {
			break
		}
		if activeSession == nil {
			sessionChan <- nil
		} else {
			sessionChan <- activeSession.Clone()
		}
	}
	if activeSession != nil {
		activeSession.Close()
	}
	close(requestChan)
	close(sessionChan)
}

//Session retrieves the current active session.
func Session() (s *mgo.Session, err error) {
	requestChan <- true
	s, ok := <-sessionChan
	if s == nil || !ok {
		err = fmt.Errorf("Could not retrieve session.")
	}
	return
}

//Close shuts down the current session.
func Close() {
	requestChan <- false
}

//DeleteDB removes a db.
func DeleteDB(db string) error {
	session, err := Session()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB(db).DropDatabase()
}

//CopyDB is used to copy the contents of one database to a new
//location.
func CopyDB(from, to string) (err error) {
	err = DeleteDB(to)
	if err != nil {
		return
	}
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	err = session.Run(
		bson.D{
			{"copydb", "1"},
			{"fromdb", from},
			{"todb", to},
		}, nil)
	return
}

//CloneCollection
func CloneCollection(origin, collection string) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	fmt.Println(session.DB("").Name, collection)
	err = session.DB("").Run(
		bson.D{
			{"cloneCollection", collection},
			{"from", origin},
		}, nil)
	return
}

//CloneData
func CloneData(origin string) (err error) {
	collections := []string{USERS, PROJECTS, SUBMISSIONS, FILES, TESTS, JPF, PMD}
	for _, col := range collections {
		err = CloneCollection(origin, col)
		if err != nil {
			return
		}
	}
	return
}

//Add adds a document to the specified collection.
func Add(colName string, data interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(colName)
	err = col.Insert(data)
	if err != nil {
		err = &DBAddError{colName, err}
	}
	return
}

//RemoveById removes a document matching the given id in collection colName from the active database.
func RemoveById(colName string, id interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	err = session.DB("").C(colName).RemoveId(id)
	if err != nil {
		err = &DBRemoveError{colName, err, id}
	}
	return
}

//Update updates documents from the collection col
//matching matcher with the changes specified by change.
func Update(col string, matcher, change interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	tcol := session.DB("").C(col)
	err = tcol.Update(matcher, change)
	if err != nil {
		err = fmt.Errorf(
			"Encountered error %q when updating %q matching %q to %q in db",
			err, col, matcher, change,
		)
	}
	return
}

//Contains checks whether the collection col contains any items matching matcher.
func Contains(col string, matcher interface{}) bool {
	n, err := Count(col, matcher)
	return err == nil && n > 0
}

//Count calculates the amount of items in the collection col which match matcher.
func Count(col string, matcher interface{}) (n int, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(col)
	n, err = c.Find(matcher).Count()
	if err != nil {
		err = &DBGetError{col + " count", err, matcher}
	}
	return
}
