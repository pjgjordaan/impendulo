package db

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/user"
)

const (
	DB          = "impendulo"
	USERS       = "users"
	SUBMISSIONS = "submissions"
	FILES       = "files"
	TOOLS       = "tools"
	ADDRESS     = "localhost"
	RESULTS     = "results"
	SET = "$set"
)

var activeSession *mgo.Session

type SingleGet func(col string, matcher interface{})(ret bson.M, err error)

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

func RemoveFileByID(id interface{}) (err error) {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(FILES)
	err = c.RemoveId(id)
	return err
}

func GetUserById(id interface{})(ret *user.User, err error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(USERS)
	err = c.FindId(id).One(&ret)
	return ret, err
}


func GetFile(matcher interface{})(ret *submission.File, err error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(FILES)
	err = c.Find(matcher).One(&ret)
	return ret, err
}

func GetSubmission(matcher interface{})(ret *submission.Submission, err error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(SUBMISSIONS)
	err = c.Find(matcher).One(&ret)
	return ret, err
}

func GetTool(matcher interface{})(ret *tool.Tool, err error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(TOOLS)
	err = c.Find(matcher).One(&ret)
	return ret, err
}


func GetTools(matcher interface{}) (ret []*tool.Tool, err error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(TOOLS)
	err = tcol.Find(matcher).All(&ret)
	return ret, err
}


func AddFile(f *submission.File) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(FILES)
	err := col.Insert(f)
	return err
}

func AddSubmission(s *submission.Submission) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(SUBMISSIONS)
	err := col.Insert(s)
	return err
}

func AddTool(t *tool.Tool) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(TOOLS)
	err := col.Insert(t)
	return err
}

func AddResult(r *tool.Result) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(RESULTS)
	err := col.Insert(r)
	return err
}


func Update(col string, matcher, change interface{}) (err error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(col)
	err = tcol.Update(matcher, change)
	return err
}


func AddUsers(users ...*user.User) (error) {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(USERS)
	return c.Insert(users)
}
