package db

import (
	"fmt"
	"labix.org/v2/mgo"
)

const (
	USERS        = "users"
	SUBMISSIONS  = "submissions"
	FILES        = "files"
	RESULTS      = "results"
	TESTS        = "tests"
	PROJECTS     = "projects"
	JPF     = "jpf"
	SET          = "$set"
	DEFAULT_CONN = "mongodb://localhost/impendulo"
	TEST_CONN    = "mongodb://localhost/impendulo_test"
	TEST_DB      = "impendulo_test"
)

var activeSession *mgo.Session

//Setup creates a mongodb session.
//This must be called before using any other db functions.
func Setup(conn string) {
	var err error
	activeSession, err = mgo.Dial(conn)
	if err != nil {
		panic(err)
	}
}

//getSession retrieves the current active session.
func getSession() (s *mgo.Session) {
	if activeSession == nil {
		panic(fmt.Errorf("Could not retrieve session."))
	}
	return activeSession.Clone()
}

func Close() {
	if activeSession != nil {
		activeSession.Close()
	}
}

func DeleteDB(db string) error {
	session := getSession()
	defer session.Close()
	return session.DB(db).DropDatabase()
}

//Update updates documents from the collection col matching the matcher interface to the change interface.
func Update(col string, matcher, change interface{}) error {
	session := getSession()
	defer session.Close()
	tcol := session.DB("").C(col)
	err := tcol.Update(matcher, change)
	if err != nil {
		return fmt.Errorf("Encountered error %q when updating %q matching %q to %q in db", err, col, matcher, change)
	}
	return nil
}

func Count(col string, matcher interface{}) (int, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(col)
	n, err := c.Find(matcher).Count()
	if err != nil {
		return -1, fmt.Errorf("Encountered error %q when counting documents matching %q in db", err, matcher)
	}
	return n, nil
}
