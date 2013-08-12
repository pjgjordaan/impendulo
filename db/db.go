package db

import (
	"fmt"
	"labix.org/v2/mgo"
	"sync"
)

const (
	USERS        = "users"
	SUBMISSIONS  = "submissions"
	FILES        = "files"
	RESULTS      = "results"
	TESTS        = "tests"
	PROJECTS     = "projects"
	JPF          = "jpf"
	SET          = "$set"
	DEFAULT_CONN = "mongodb://localhost/impendulo"
	TEST_CONN    = "mongodb://localhost/impendulo_test"
	TEST_DB      = "impendulo_test"
)

var activeSession *mgo.Session
var dbLock *sync.Mutex

func init() {
	dbLock = new(sync.Mutex)
}

//Setup creates a mongodb session.
//This must be called before using any other db functions.
func Setup(conn string) (err error) {
	dbLock.Lock()
	activeSession, err = mgo.Dial(conn)
	dbLock.Unlock()
	return
}

//getSession retrieves the current active session.
func getSession() (s *mgo.Session, err error) {
	dbLock.Lock()
	if activeSession == nil {
		err = fmt.Errorf("Could not retrieve session.")
	} else {
		s = activeSession.Clone()
	}
	dbLock.Unlock()
	return
}

//Close shuts down the current session.
func Close() {
	dbLock.Lock()
	if activeSession != nil {
		activeSession.Close()
	}
	dbLock.Unlock()
}

//DeleteDB removes a db.
func DeleteDB(db string) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB(db).DropDatabase()
}

//Update updates documents from the collection col 
//matching the matcher interface to the change interface.
func Update(col string, matcher, change interface{}) (err error) {
	session, err := getSession()
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

//Contains checks whether the collection col 
//contains any items matching the interfacce{} matcher.
func Contains(col string, matcher interface{}) bool {
	n, err := Count(col, matcher)
	return err == nil && n > 0
}

//Count calculates the amount of items in the collection col 
//which match the interface{} matcher. 
func Count(col string, matcher interface{}) (n int, err error) {
	session, err := getSession()
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

//DBGetError represents errors encountered 
//when retrieving data from the db.
type DBGetError struct {
	tipe    string
	err     error
	matcher interface{}
}

func (this *DBGetError) Error() string {
	return fmt.Sprintf(
		"Encountered error %q when retrieving %q matching %q from db", 
		this.err, this.tipe, this.matcher,
	)
}

//DBAddError represents errors encountered 
//when adding data to the db.
type DBAddError struct {
	msg string
	err error
}

func (this *DBAddError) Error() string {
	return fmt.Sprintf(
		"Encountered error %q when adding %q to db", 
		this.err, this.msg,
	)
}

//DBRemoveError represents errors encountered 
//when removing data from the db.
type DBRemoveError struct {
	tipe    string
	err     error
	matcher interface{}
}

func (this *DBRemoveError) Error() string {
	return fmt.Sprintf(
		"Encountered error %q when removing %q matching %q from db", 
		this.err, this.tipe, this.matcher,
	)
}
