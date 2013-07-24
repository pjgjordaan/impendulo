package db

import (
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
)

//GetFile retrieves a file matching the given interface from the active database.
func GetFile(matcher, selector interface{}) (ret *project.File, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"file", err, matcher}
	}
	return
}

//GetFiles retrieves files matching the given interface from the active database.
func GetFiles(matcher, selector interface{}, sort string) (ret []*project.File, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	q := c.Find(matcher)
	if sort != "" {
		q = q.Sort(sort)
	}
	err = q.Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"files", err, matcher}
	}
	return
}

func GetFileNames(matcher interface{}) (ret []string, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.Find(matcher).Distinct(project.NAME, &ret)
	if err != nil {
		err = &DBGetError{"filenames", err, matcher}
	}
	return
}

//GetSubmission retrieves a submission matching the given interface from the active database.
func GetSubmission(matcher, selector interface{}) (ret *project.Submission, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"submission", err, matcher}
	}
	return
}

//GetSubmission retrieves submissions matching the given interface from the active database.
func GetSubmissions(matcher, selector interface{}) (ret []*project.Submission, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	err = c.Find(matcher).Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"submissions", err, matcher}
	}
	return
}

//GetTest retrieves a test matching the given interface from the active database.
func GetTest(matcher, selector interface{}) (ret *project.Test, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TESTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"test", err, matcher}
	}
	return
}

//GetTest retrieves a test matching the given interface from the active database.
func GetTests(matcher, selector interface{}) (ret []*project.Test, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TESTS)
	err = c.Find(matcher).Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"tests", err, matcher}
	}
	return
}

func GetJPF(matcher, selector interface{}) (ret *project.JPFFile, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(JPF)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"jpf config file", err, matcher}
	}
	return
}

//GetTest retrieves a test matching the given interface from the active database.
func GetProject(matcher, selector interface{}) (ret *project.Project, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"project", err, matcher}
	}
	return
}

//GetTest retrieves a test matching the given interface from the active database.
	func GetProjects(matcher interface{}) (ret []*project.Project, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	err = c.Find(matcher).Select(bson.M{project.SKELETON:0}).All(&ret)
	if err != nil {
		err = &DBGetError{"projects", err, matcher}
	}
	return
}

//AddFile adds a new file to the active database.
func AddFile(f *project.File) (err error) {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(FILES)
	err = col.Insert(f)
	if err != nil {
		err = &DBAddError{f.String(), err}
	}
	return
}

//AddFile adds a new file to the active database.
func AddTest(t *project.Test) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(TESTS)
	matcher := bson.M{project.PROJECT_ID: t.ProjectId}
	_, err := col.RemoveAll(matcher)
	if err != nil {
		err = &DBRemoveError{"test files", err, matcher}
	}
	err = col.Insert(t)
	if err != nil {
		err = &DBAddError{t.String(), err}
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
		err = &DBRemoveError{"jpf config files", err, matcher}
	}
	err = col.Insert(jpf)
	if err != nil {
		err = &DBAddError{jpf.String(), err}
	}
	return nil
}

//AddSubmission adds a new submission to the active database.
func AddSubmission(s *project.Submission) (err error) {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(SUBMISSIONS)
	err = col.Insert(s)
	if err != nil {
		err = &DBAddError{s.String(), err}
	}
	return
}

//AddSubmission adds a new submission to the active database.
func AddProject(p *project.Project) (err error) {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(PROJECTS)
	err = col.Insert(p)
	if err != nil {
		err = &DBAddError{p.String(), err}
	}
	return
}

//RemoveFileById removes a file matching the given id from the active database.
func RemoveFileByID(id interface{}) (err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"file", err, id}
	}
	return
}
