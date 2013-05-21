package db

import (
	"fmt"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/user"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	USERS        = "users"
	SUBMISSIONS  = "submissions"
	FILES        = "files"
	TOOLS        = "tools"
	RESULTS      = "results"
	TESTS = "tests"
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

func DeleteDB(db string) error {
	session := getSession()
	defer session.Close()
	return session.DB(db).DropDatabase()
}

//RemoveFileById removes a file matching the given id from the active database.
func RemoveFileByID(id interface{}) error {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	err := c.RemoveId(id)
	if err != nil {
		return fmt.Errorf("Encountered error %q when removing file %q from db", err, id)
	}
	return nil
}

//GetUserById retrieves a user matching the given id from the active database. 
func GetUserById(id interface{}) (*user.User, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	var ret *user.User
	err := c.FindId(id).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving user %q from db", err, id)
	}
	return ret, nil
}

//GetFile retrieves a file matching the given interface from the active database. 
func GetFile(matcher interface{}) (*submission.File, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	var ret *submission.File
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving file matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetSubmission retrieves a submission matching the given interface from the active database.
func GetSubmission(matcher interface{}) (*submission.Submission, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	var ret *submission.Submission
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving submission matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTool retrieves a tool matching the given interface from the active database.
func GetTool(matcher interface{}) (*tool.Tool, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TOOLS)
	var ret *tool.Tool
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving tool matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTools retrieves tools matching the given interface from the active database.
func GetTools(matcher interface{}) ([]*tool.Tool, error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB("").C(TOOLS)
	var ret []*tool.Tool
	err := tcol.Find(matcher).All(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving tools matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetResult retrieves a result matching the given interface from the active database.
func GetResult(matcher interface{}) (*tool.Result, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	var ret *tool.Result
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving result matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTest retrieves a test matching the given interface from the active database.
func GetTest(matcher interface{}) (*submission.Test, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TESTS)
	var ret *submission.Test
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving test matching %q from db", err, matcher)
	}
	return ret, nil
}


//AddFile adds a new file to the active database.
func AddFile(f *submission.File) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(FILES)
	err := col.Insert(f)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding file %q to db", err, f)
	}
	return nil
}

//AddFile adds a new file to the active database.
func AddTest(t *submission.Test) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(TESTS)
	matcher := bson.M{submission.PROJECT: t.Project}
	_, err := col.RemoveAll(matcher)
	if err != nil {
		return fmt.Errorf("Encountered error %q when removing tests matching %q to db", err, matcher)
	}
	err = col.Insert(t)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding test %q to db", err, t)
	}
	return nil
}


//AddSubmission adds a new submission to the active database.
func AddSubmission(s *submission.Submission) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(SUBMISSIONS)
	err := col.Insert(s)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding submission %q to db", err, s)
	}
	return nil
}

//AddTool adds a new tool to the active database.
func AddTool(t *tool.Tool) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(TOOLS)
	err := col.Insert(t)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding tool %q to db", err, t)
	}
	return nil
}

//AddResult adds a new result to the active database.
func AddResult(r *tool.Result) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err := col.Insert(r)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding result %q to db", err, r)
	}
	return nil
}

//AddUser adds a new user to the active database.
func AddUser(u *user.User) error {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err := c.Insert(u)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding user %q to db", err, u)
	}
	return nil
}

//AddUsers adds new users to the active database.
func AddUsers(users ...*user.User) error {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(USERS)
	err := c.Insert(users)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding users %q to db", err, users)
	}
	return nil
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
