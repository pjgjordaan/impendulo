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

package db

import (
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
)

type (
	//FileInfo
	FileInfo struct {
		Name  string
		Count int
	}
)

//File retrieves a file matching the given interface from the active database.
func File(matcher, selector interface{}) (ret *project.File, err error) {
	session, err := Session()
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

//Files retrieves files matching the given interface from the active database.
func Files(matcher, selector interface{}, sort ...string) (ret []*project.File, err error) {
	session, err := Session()
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

//FileInfos retrieves names of file information.
func FileInfos(matcher bson.M) (ret []*FileInfo, err error) {
	names, err := FileNames(matcher)
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

//FileNames retrieves names of files
//matching the given interface from the active database.
func FileNames(matcher interface{}) (ret []string, err error) {
	session, err := Session()
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

//Submission retrieves a submission matching
//the given interface from the active database.
func Submission(matcher, selector interface{}) (ret *project.Submission, err error) {
	session, err := Session()
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

//Submissions retrieves submissions matching
//the given interface from the active database.
func Submissions(matcher, selector interface{}, sort ...string) (ret []*project.Submission, err error) {
	session, err := Session()
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

//Project retrieves a project matching
//the given interface from the active database.
func Project(matcher, selector interface{}) (ret *project.Project, err error) {
	session, err := Session()
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

//Projects retrieves projects matching
//the given interface from the active database.
func Projects(matcher, selector interface{}, sort ...string) (ret []*project.Project, err error) {
	session, err := Session()
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

//RemoveFileById removes a file matching the given id from the active database.
func RemoveFileById(id interface{}) (err error) {
	file, err := File(bson.M{project.ID: id}, bson.M{project.RESULTS: 1})
	if err != nil {
		return
	}
	for _, resId := range file.Results {
		if _, ok := resId.(bson.ObjectId); !ok {
			continue
		}
		err = RemoveById(RESULTS, resId)
		if err != nil {
			return
		}
	}
	err = RemoveById(FILES, id)
	return
}

//RemoveSubmissionById removes a submission matching
//the given id from the active database.
func RemoveSubmissionById(id interface{}) (err error) {
	files, err := Files(bson.M{project.SUBID: id},
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
	err = RemoveById(SUBMISSIONS, id)
	return
}

//RemoveProjectById removes a project matching
//the given id from the active database.
func RemoveProjectById(id interface{}) (err error) {
	projectMatch := bson.M{project.PROJECT_ID: id}
	idSelect := bson.M{project.ID: 1}
	subs, err := Submissions(projectMatch, idSelect)
	if err != nil {
		return
	}
	for _, sub := range subs {
		err = RemoveSubmissionById(sub.Id)
		if err != nil {
			return
		}
	}
	tests, err := JUnitTests(projectMatch, idSelect)
	if err != nil {
		return
	}
	for _, test := range tests {
		RemoveById(TESTS, test.Id)
	}
	jpfConfig, err := JPFConfig(projectMatch, idSelect)
	if err == nil {
		RemoveById(JPF, jpfConfig.Id)
	}
	pmdRules, err := PMDRules(projectMatch, idSelect)
	if err == nil {
		RemoveById(PMD, pmdRules.Id)
	}
	err = RemoveById(PROJECTS, id)
	return
}
