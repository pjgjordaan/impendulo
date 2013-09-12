package db

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
)

//GetJPF retrieves a JPF configuration file
//matching the given interface from the active database.
func GetJPF(matcher, selector interface{}) (ret *jpf.Config, err error) {
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

//AddJPF adds a new JPF configuration file to the active database.
func AddJPF(cfg *jpf.Config) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	col := session.DB("").C(JPF)
	matcher := bson.M{project.PROJECT_ID: cfg.ProjectId}
	_, err = col.RemoveAll(matcher)
	if err != nil {
		err = &DBRemoveError{"jpf config files", err, matcher}
	}
	err = col.Insert(cfg)
	if err != nil {
		err = &DBAddError{cfg.String(), err}
	}
	return nil
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

func GetPMD(matcher, selector interface{}) (ret *pmd.Rules, err error) {
	session, err := getSession()
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

func AddPMD(rules *pmd.Rules) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	col := session.DB("").C(PMD)
	matcher := bson.M{project.PROJECT_ID: rules.ProjectId}
	_, err = col.RemoveAll(matcher)
	if err != nil {
		err = &DBRemoveError{"pmd rules", err, matcher}
	}
	err = col.Insert(rules)
	if err != nil {
		err = &DBAddError{"pmd rules", err}
	}
	return nil
}

func RemovePMDById(id interface{}) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PMD)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"pmd", err, id}
	}
	return
}

//GetTest retrieves a test matching the given interface from the active database.
func GetTest(matcher, selector interface{}) (ret *junit.Test, err error) {
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
func GetTests(matcher, selector interface{}) (ret []*junit.Test, err error) {
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

//AddTest adds a new test to the active database.
func AddTest(t *junit.Test) error {
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
