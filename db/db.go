package db

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	USERS        = "users"
	SUBMISSIONS  = "submissions"
	FILES        = "files"
	RESULTS      = "results"
	TESTS        = "tests"
	PROJECTS     = "projects"
	JPF          = "jpf"
	PMD          = "pmd"
	SET          = "$set"
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

//getSession retrieves the current active session.
func getSession() (s *mgo.Session, err error) {
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
	session, err := getSession()
	if err != nil {
		return err
	}
	return session.DB(db).DropDatabase()
}

//CopyDB is used to copy the contents of one database to a new
//location.
func CopyDB(from, to string) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	err = DeleteDB(to)
	if err != nil {
		return err
	}
	return session.Run(bson.D{
		{"copydb", "1"}, {"fromdb", from},
		{"todb", to}}, nil)
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
