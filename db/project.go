package db

import (
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
)

//GetFile retrieves a file matching the given interface from the active database.
func GetFile(matcher, selector interface{}) (ret *project.File, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
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
	session, err := getSession()
	if err != nil {
		return
	}
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

//GetFileNames retrieves names of files
//matching the given interface from the active database.
func GetFileNames(matcher interface{}) (ret []string, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.Find(matcher).Distinct(project.NAME, &ret)
	if err != nil {
		err = &DBGetError{"filenames", err, matcher}
	}
	return
}

//GetSubmission retrieves a submission matching
//the given interface from the active database.
func GetSubmission(matcher, selector interface{}) (ret *project.Submission, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"submission", err, matcher}
	}
	return
}

//GetSubmissions retrieves submissions matching
//the given interface from the active database.
func GetSubmissions(matcher, selector interface{}) (ret []*project.Submission, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
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
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(TESTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"test", err, matcher}
	}
	return
}

//GetTest retrieves tests matching
//the given interface from the active database.
func GetTests(matcher, selector interface{}) (ret []*project.Test, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(TESTS)
	err = c.Find(matcher).Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"tests", err, matcher}
	}
	return
}

//GetJPF retrieves a JPF configuration file
//matching the given interface from the active database.
func GetJPF(matcher, selector interface{}) (ret *project.JPFFile, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(JPF)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"jpf config file", err, matcher}
	}
	return
}

//GetProject retrieves a project matching
//the given interface from the active database.
func GetProject(matcher, selector interface{}) (ret *project.Project, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"project", err, matcher}
	}
	return
}

//GetProjects retrieves projects matching
//the given interface from the active database.
func GetProjects(matcher interface{}) (ret []*project.Project, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	err = c.Find(matcher).Select(bson.M{project.SKELETON: 0}).All(&ret)
	if err != nil {
		err = &DBGetError{"projects", err, matcher}
	}
	return
}

//AddFile adds a new file to the active database.
func AddFile(f *project.File) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(FILES)
	err = col.Insert(f)
	if err != nil {
		err = &DBAddError{f.String(), err}
	}
	return
}

//AddTest adds a new test to the active database.
func AddTest(t *project.Test) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	col := session.DB("").C(TESTS)
	err = col.Insert(t)
	if err != nil {
		err = &DBAddError{t.String(), err}
	}
	return nil
}

//AddJPF adds a new JPF configuration file to the active database.
func AddJPF(jpf *project.JPFFile) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	col := session.DB("").C(JPF)
	matcher := bson.M{project.PROJECT_ID: jpf.ProjectId}
	_, err = col.RemoveAll(matcher)
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
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(SUBMISSIONS)
	err = col.Insert(s)
	if err != nil {
		err = &DBAddError{s.String(), err}
	}
	return
}

//AddProject adds a new project to the active database.
func AddProject(p *project.Project) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(PROJECTS)
	err = col.Insert(p)
	if err != nil {
		err = &DBAddError{p.String(), err}
	}
	return
}

//RemoveFileById removes a file matching the given id from the active database.
func RemoveFileById(id interface{}) (err error) {
	file, err := GetFile(bson.M{project.ID: id}, bson.M{project.RESULTS: 1})
	if err != nil {
		return
	}
	for _, resId := range file.Results {
		err = RemoveResultById(resId)
		if err != nil {
			return
		}
	}
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"file", err, id}
	}
	return
}

//RemoveSubmissionById removes a submission matching
//the given id from the active database.
func RemoveSubmissionById(id interface{}) (err error) {
	files, err := GetFiles(bson.M{project.SUBID: id},
		bson.M{project.ID: 1}, "")
	if err != nil {
		return
	}
	for _, file := range files {
		err = RemoveFileById(file.Id)
		if err != nil {
			return
		}
	}
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"submission", err, id}
	}
	return
}

//RemoveJPFById removes a JPF configuration file matching
//the given id from the active database.
func RemoveJPFById(id interface{}) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(JPF)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"jpf", err, id}
	}
	return
}

//RemoveTestById removes a test matching the given id from the active database.
func RemoveTestById(id interface{}) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(TESTS)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"test", err, id}
	}
	return
}

//RemoveProjectById removes a project matching
//the given id from the active database.
func RemoveProjectById(id interface{}) (err error) {
	projectMatch := bson.M{project.PROJECT_ID: id}
	idSelect := bson.M{project.ID: 1}
	subs, err := GetSubmissions(projectMatch, idSelect)
	if err != nil {
		return
	}
	for _, sub := range subs {
		err = RemoveSubmissionById(sub.Id)
		if err != nil {
			return
		}
	}
	tests, err := GetTests(projectMatch, idSelect)
	if err != nil {
		return
	}
	for _, test := range tests {
		err = RemoveTestById(test.Id)
		if err != nil {
			return
		}
	}
	jpfConfig, err := GetJPF(projectMatch, idSelect)
	if err == nil {
		RemoveJPFById(jpfConfig.Id)
	}
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"project", err, id}
	}
	return
}
