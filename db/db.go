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
	DEFAULT_CONN = "mongodb://localhost/impendulo"
	DEBUG_CONN   = "mongodb://localhost/impendulo_debug"
	TEST_CONN    = "mongodb://localhost/impendulo_test"
	DEFAULT_DB   = "impendulo"
	DEBUG_DB     = "impendulo_debug"
	TEST_DB      = "impendulo_test"
	BACKUP_DB    = "impendulo_backup"
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
