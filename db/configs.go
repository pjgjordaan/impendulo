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
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
)

//JPFConfig retrieves a JPF configuration matching matcher from the active database.
func JPFConfig(matcher, selector interface{}) (ret *jpf.Config, err error) {
	session, err := Session()
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

//AddJPF overwrites a project's JPF configuration with the provided configuration.
func AddJPFConfig(cfg *jpf.Config) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(JPF)
	matcher := bson.M{project.PROJECT_ID: cfg.ProjectId}
	col.RemoveAll(matcher)
	err = col.Insert(cfg)
	if err != nil {
		err = &DBAddError{cfg.String(), err}
	}
	return
}

//PMDRules retrieves PMD rules matching matcher from the db.
func PMDRules(matcher, selector interface{}) (ret *pmd.Rules, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PMD)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"pmd rules", err, matcher}
	}
	return
}

//AddPMDRules overwrites a project's current PMD rules with the provided rules.
func AddPMDRules(rules *pmd.Rules) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(PMD)
	matcher := bson.M{project.PROJECT_ID: rules.ProjectId}
	col.RemoveAll(matcher)
	err = col.Insert(rules)
	if err != nil {
		err = &DBAddError{"pmd rules", err}
	}
	return
}

//JUnitTest retrieves a test matching the matcher from the active database.
func JUnitTest(matcher, selector interface{}) (ret *junit.Test, err error) {
	session, err := Session()
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

//JUnitTests retrieves all tests matching matcher from the active database.
func JUnitTests(matcher, selector interface{}) (ret []*junit.Test, err error) {
	session, err := Session()
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

//AddJUnitTest overwrites one of a project's JUnit tests with the new JUnit test
//if it has the same name as the new test. Otherwise the new test is just added to the project's tests.
func AddJUnitTest(t *junit.Test) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(TESTS)
	matcher := bson.M{project.PROJECT_ID: t.ProjectId, project.NAME: t.Name}
	col.RemoveAll(matcher)
	err = col.Insert(t)
	if err != nil {
		err = &DBAddError{t.String(), err}
	}
	return
}
