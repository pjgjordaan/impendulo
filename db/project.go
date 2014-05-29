//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package db

import (
	"fmt"

	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
)

type (
	//FileInfo
	FileInfo struct {
		Name  string
		Count int
		Type  project.Type
	}
)

func (f *FileInfo) HasCharts() bool {
	return f.Type == project.SRC
}

//File retrieves a file matching m from the active database.
func File(m, sl interface{}) (*project.File, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var f *project.File
	if e = s.DB("").C(FILES).Find(m).Select(sl).One(&f); e != nil {
		return nil, &GetError{"file", e, m}
	}
	return f, nil
}

//Files retrieves files matching m from the active database.
func Files(m, sl interface{}, limit int, sort ...string) ([]*project.File, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	q := s.DB("").C(FILES).Find(m).Limit(limit)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	var fs []*project.File
	if e = q.Select(sl).All(&fs); e != nil {
		return nil, &GetError{"files", e, m}
	}
	return fs, nil
}

//FileInfos retrieves names of file information.
func FileInfos(m bson.M) ([]*FileInfo, error) {
	ns, e := FileNames(m)
	if e != nil {
		return nil, e
	}
	fi := make([]*FileInfo, len(ns))
	for i, n := range ns {
		m[NAME] = n
		c, e := Count(FILES, m)
		if e != nil {
			return nil, e
		}
		f, e := File(m, bson.M{TYPE: 1})
		if e != nil {
			return nil, e
		}
		fi[i] = &FileInfo{Name: n, Count: c, Type: f.Type}
	}
	return fi, nil
}

//FileNames retrieves names of files
//matching the given interface from the active database.
func FileNames(m interface{}) ([]string, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var ns []string
	if e = s.DB("").C(FILES).Find(m).Distinct(NAME, &ns); e != nil {
		return nil, &GetError{"filenames", e, m}
	}
	return ns, nil
}

//Submission retrieves a submission matching m from the active database.
func Submission(m, sl interface{}) (*project.Submission, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var sb *project.Submission
	if e = s.DB("").C(SUBMISSIONS).Find(m).Select(sl).One(&sb); e != nil {
		return nil, &GetError{"submission", e, m}
	}
	return sb, nil
}

//Submissions retrieves submissions matching m from the active database.
func Submissions(m, sl interface{}, sort ...string) ([]*project.Submission, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	q := s.DB("").C(SUBMISSIONS).Find(m)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	var ss []*project.Submission
	if e = q.Select(sl).All(&ss); e != nil {
		return nil, &GetError{"submissions", e, m}
	}
	return ss, nil
}

//Project retrieves a project matching m from the active database.
func Project(m, sl interface{}) (*project.Project, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var p *project.Project
	if e = s.DB("").C(PROJECTS).Find(m).Select(sl).One(&p); e != nil {
		return nil, &GetError{"project", e, m}
	}
	return p, nil
}

//Projects retrieves projects matching m from the active database.
func Projects(m, sl interface{}, sort ...string) ([]*project.Project, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	q := s.DB("").C(PROJECTS).Find(m)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	var p []*project.Project
	if e = q.Select(sl).All(&p); e != nil {
		return nil, &GetError{"projects", e, m}
	}
	return p, nil
}

func Skeleton(m, sl interface{}) (*project.Skeleton, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var sk *project.Skeleton
	if e = s.DB("").C(SKELETONS).Find(m).Select(sl).One(&sk); e != nil {
		return nil, &GetError{"skeleton", e, m}
	}
	return sk, nil
}

func Skeletons(m, sl interface{}, sort ...string) ([]*project.Skeleton, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	q := s.DB("").C(SKELETONS).Find(m)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	var sk []*project.Skeleton
	if e = q.Select(sl).All(&sk); e != nil {
		return nil, &GetError{"skeletons", e, m}
	}
	return sk, nil
}

//RemoveFileById removes a file matching the given id from the active database.
func RemoveFileById(id interface{}) error {
	f, e := File(bson.M{ID: id}, bson.M{RESULTS: 1})
	if e != nil {
		return e
	}
	for _, r := range f.Results {
		if _, ok := r.(bson.ObjectId); !ok {
			continue
		}
		RemoveById(RESULTS, r)
	}
	return RemoveById(FILES, id)
}

//RemoveSubmissionById removes a submission matching
//the given id from the active database.
func RemoveSubmissionById(id interface{}) error {
	fs, e := Files(bson.M{SUBID: id}, bson.M{ID: 1}, 0)
	if e != nil {
		return e
	}
	for _, f := range fs {
		if e = RemoveFileById(f.Id); e != nil {
			return e
		}
	}
	return RemoveById(SUBMISSIONS, id)
}

//RemoveProjectById removes a project matching
//the given id from the active database.
func RemoveProjectById(id interface{}) error {
	pm := bson.M{PROJECTID: id}
	is := bson.M{ID: 1}
	ss, e := Submissions(pm, is)
	if e != nil {
		return e
	}
	for _, s := range ss {
		if e = RemoveSubmissionById(s.Id); e != nil {
			return e
		}
	}
	sks, e := Skeletons(pm, is)
	if e != nil {
		return e
	}
	for _, sk := range sks {
		RemoveById(SKELETONS, sk.Id)
	}
	ts, e := JUnitTests(pm, is)
	if e != nil {
		return e
	}
	for _, t := range ts {
		RemoveById(TESTS, t.Id)
	}
	c, e := JPFConfig(pm, is)
	if e == nil {
		RemoveById(JPF, c.Id)
	}
	r, e := PMDRules(pm, is)
	if e == nil {
		RemoveById(PMD, r.Id)
	}
	return RemoveById(PROJECTS, id)
}

func LastFile(m, sl interface{}) (*project.File, error) {
	return firstFile(m, sl, "-"+TIME)
}

func FirstFile(m, sl interface{}) (*project.File, error) {
	return firstFile(m, sl, TIME)
}

func firstFile(m, sl interface{}, sort string) (*project.File, error) {
	fs, e := Files(m, sl, 0, sort)
	if e != nil {
		return nil, e
	}
	if len(fs) == 0 {
		return nil, fmt.Errorf("no files for matcher %q", m)
	}
	return fs[0], nil
}

func UpdateTime(sub *project.Submission) error {
	f, e := FirstFile(bson.M{SUBID: sub.Id}, bson.M{TIME: 1})
	if e != nil {
		return e
	}
	if f.Time >= sub.Time {
		return nil
	}
	return Update(SUBMISSIONS, bson.M{ID: sub.Id}, bson.M{SET: bson.M{TIME: f.Time}})
}
