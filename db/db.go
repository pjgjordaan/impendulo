package db

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	DB          = "impendulo"
	USERS       = "users"
	SUBMISSIONS = "submissions"
	FILES       = "files"
	TOOLS       = "tools"
	ADDRESS     = "localhost"
)

var activeSession *mgo.Session

func getSession() (s *mgo.Session) {
	if activeSession == nil {
		var err error
		activeSession, err = mgo.Dial(ADDRESS)
		if err != nil {
			panic(err)
		}
	}
	s = activeSession.Clone()
	return s
}

func GetById(col string, id interface{}) (ret bson.M, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(col)
	err = c.FindId(id).One(&ret)
	return ret, err
}

func GetAll(col string, matcher interface{}) (ret []bson.M, err error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(col)
	err = tcol.Find(matcher).All(&ret)
	return ret, err
}

func GetMatching(col string, matcher interface{}) (ret bson.M, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(col)
	err = c.Find(matcher).One(&ret)
	return ret, err
}

func AddSingle(col string, item interface{}) (err error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(col)
	err = tcol.Insert(item)
	return err
}

func AddMany(col string, items ...interface{}) (err error) {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(col)
	for _, item := range items {
		err = c.Insert(item)
		if err != nil {
			break
		}
	}
	return err
}

func Update(col string, matcher, change interface{}) (err error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(col)
	err = tcol.Update(matcher, change)
	return err
}

func UpsertId(col string, id, item interface{}) (err error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(col)
	_, err = tcol.UpsertId(id, item)
	return err
}
