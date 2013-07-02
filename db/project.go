package db

import (
	"fmt"
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
)

//GetFile retrieves a file matching the given interface from the active database.
func GetFile(matcher, selector interface{}) (*project.File, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	var ret *project.File
	err := c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving file matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetFiles retrieves files matching the given interface from the active database.
func GetFiles(matcher, selector interface{}) ([]*project.File, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	var ret []*project.File
	err := c.Find(matcher).Select(selector).All(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving files matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetSubmission retrieves a submission matching the given interface from the active database.
func GetSubmission(matcher, selector interface{}) (*project.Submission, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	var ret *project.Submission
	err := c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving submission matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetSubmission retrieves submissions matching the given interface from the active database.
func GetSubmissions(matcher, selector interface{}) ([]*project.Submission, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	var ret []*project.Submission
	err := c.Find(matcher).Select(selector).All(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving submissions matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTest retrieves a test matching the given interface from the active database.
func GetTest(matcher, selector interface{}) (*project.Test, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TESTS)
	var ret *project.Test
	err := c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving test matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTest retrieves a test matching the given interface from the active database.
func GetTests(matcher, selector interface{}) ([]*project.Test, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TESTS)
	var ret []*project.Test
	err := c.Find(matcher).Select(selector).All(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving test matching %q from db", err, matcher)
	}
	return ret, nil
}

func GetJPF(matcher, selector interface{}) (*project.JPFFile, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(JPF)
	var ret *project.JPFFile
	err := c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving jpf config matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTest retrieves a test matching the given interface from the active database.
func GetProject(matcher, selector interface{}) (*project.Project, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	var ret *project.Project
	err := c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving project matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTest retrieves a test matching the given interface from the active database.
func GetProjects(matcher, selector interface{}) ([]*project.Project, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	var ret []*project.Project
	err := c.Find(matcher).All(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving projects matching %q from db", err, matcher)
	}
	return ret, nil
}

//AddFile adds a new file to the active database.
func AddFile(f *project.File) error {
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
func AddTest(t *project.Test) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(TESTS)
	matcher := bson.M{project.PROJECT_ID: t.ProjectId}
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

func AddJPF(jpf *project.JPFFile) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(JPF)
	matcher := bson.M{project.PROJECT_ID: jpf.ProjectId}
	_, err := col.RemoveAll(matcher)
	if err != nil {
		return fmt.Errorf("Encountered error %q when removing jpf configs matching %q from db", err, matcher)
	}
	err = col.Insert(jpf)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding jpf config %q to db", err, jpf)
	}
	return nil
}

//AddSubmission adds a new submission to the active database.
func AddSubmission(s *project.Submission) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(SUBMISSIONS)
	err := col.Insert(s)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding submission %q to db", err, s)
	}
	return nil
}

//AddSubmission adds a new submission to the active database.
func AddProject(p *project.Project) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(PROJECTS)
	err := col.Insert(p)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding project %q to db", err, p)
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
