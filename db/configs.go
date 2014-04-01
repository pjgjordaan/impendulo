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
//ANY THERRORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package db

import (
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	mk "github.com/godfried/impendulo/tool/make"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
)

//JPFConfig retrieves a JPF configuration matching m from the active database.
func JPFConfig(m, sl interface{}) (*jpf.Config, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var c *jpf.Config
	if e = s.DB("").C(JPF).Find(m).Select(sl).One(&c); e != nil {
		return nil, &GetError{"jpf config file", e, m}
	}
	return c, nil
}

//AddJPF overwrites a project's JPF configuration with the provided configuration.
func AddJPFConfig(cfg *jpf.Config) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	c := s.DB("").C(JPF)
	c.RemoveAll(bson.M{PROJECTID: cfg.ProjectId})
	if e = c.Insert(cfg); e != nil {
		return &AddError{cfg.String(), e}
	}
	return nil
}

//PMDRules retrieves PMD rules matching m from the db.
func PMDRules(m, sl interface{}) (*pmd.Rules, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *pmd.Rules
	if e = s.DB("").C(PMD).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"pmd rules", e, m}
	}
	return r, nil
}

//AddPMDRules overwrites a project's current PMD rules with the provided rules.
func AddPMDRules(r *pmd.Rules) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	c := s.DB("").C(PMD)
	m := bson.M{PROJECTID: r.ProjectId}
	c.RemoveAll(m)
	if e = c.Insert(r); e != nil {
		return &AddError{"pmd rules", e}
	}
	return nil
}

//JUnitTest retrieves a test matching the m from the active database.
func JUnitTest(m, sl interface{}) (*junit.Test, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var t *junit.Test
	if e = s.DB("").C(TESTS).Find(m).Select(sl).One(&t); e != nil {
		return nil, &GetError{"test", e, m}
	}
	return t, nil
}

//JUnitTests retrieves all tests matching m from the active database.
func JUnitTests(m, sl interface{}) ([]*junit.Test, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var t []*junit.Test
	if e = s.DB("").C(TESTS).Find(m).Select(sl).All(&t); e != nil {
		return nil, &GetError{"tests", e, m}
	}
	return t, nil
}

//AddJUnitTest overwrites one of a project's JUnit tests with the new JUnit test
//if it has the same name as the new test. Otherwise the new test is just added to the project's tests.
func AddJUnitTest(t *junit.Test) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	c := s.DB("").C(TESTS)
	c.RemoveAll(bson.M{PROJECTID: t.ProjectId, NAME: t.Name})
	if e = c.Insert(t); e != nil {
		return &AddError{t.Name, e}
	}
	return nil
}

func Makefile(m, sl interface{}) (*mk.Makefile, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var mf *mk.Makefile
	if e = s.DB("").C(MAKE).Find(m).Select(sl).One(&mf); e != nil {
		return nil, &GetError{"makefile", e, m}
	}
	return mf, nil
}

func AddMakefile(mf *mk.Makefile) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	c := s.DB("").C(MAKE)
	c.RemoveAll(bson.M{PROJECTID: mf.ProjectId})
	if e = c.Insert(mf); e != nil {
		return &AddError{"makefile", e}
	}
	return nil
}
