package db

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/jpf"
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
func AddJPF(jpf *jpf.Config) error {
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
