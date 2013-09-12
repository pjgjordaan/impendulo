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
func GetFiles(matcher, selector interface{}, sort ...string) (ret []*project.File, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(FILES)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"files", err, matcher}
	}
	return
}

type FileInfo struct {
	Name  string
	Count int
}

//GetFileInfo retrieves names of file information.
func GetFileInfo(matcher bson.M) (ret []*FileInfo, err error) {
	names, err := GetFileNames(matcher)
	if err != nil {
		return
	}
	ret = make([]*FileInfo, len(names))
	for i, name := range names {
		ret[i] = new(FileInfo)
		ret[i].Name = name
		matcher[project.NAME] = name
		ret[i].Count, err = Count(FILES, matcher)
		if err != nil {
			return
		}
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
func GetSubmissions(matcher, selector interface{}, sort ...string) (ret []*project.Submission, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"submissions", err, matcher}
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
func GetProjects(matcher, selector interface{}, sort ...string) (ret []*project.Project, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(selector).All(&ret)
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
		if _, ok := resId.(bson.ObjectId); !ok {
			continue
		}
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
		bson.M{project.ID: 1})
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
	pmdRules, err := GetPMD(projectMatch, idSelect)
	if err == nil {
		RemovePMDById(pmdRules.Id)
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
