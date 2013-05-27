package db

import(
	"github.com/godfried/cabanga/submission"
"fmt"
	"labix.org/v2/mgo/bson"
)

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
