package db

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/user"
	"fmt"
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

func RemoveFileByID(id interface{}) error {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(FILES)
	err := c.RemoveId(id)
	if err != nil{
		return fmt.Errorf("Encountered error %q when removing file %q from db", err, id) 
	}
	return nil
}

func GetUserById(id interface{})(*user.User, error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(USERS)
	var ret *user.User
	err := c.FindId(id).One(&ret)
	if err != nil{
		return nil, fmt.Errorf("Encountered error %q when retrieving user %q from db", err, id)
	}
	return ret, nil
}


func GetFile(matcher interface{})(*submission.File, error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(FILES)
	var ret *submission.File
	err := c.Find(matcher).One(&ret)
	if err != nil{
		return nil, fmt.Errorf("Encountered error %q when retrieving file matching %q from db", err, matcher)
	}
	return ret, nil
}


func GetSubmission(matcher interface{})(*submission.Submission, error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(SUBMISSIONS)
	var ret *submission.Submission
	err := c.Find(matcher).One(&ret)
	if err != nil{
		return nil, fmt.Errorf("Encountered error %q when retrieving submission matching %q from db", err, matcher)
	}
	return ret, nil
}

func GetTool(matcher interface{})(*tool.Tool, error){
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(TOOLS)
	var ret *tool.Tool
	err := c.Find(matcher).One(&ret)
	if err != nil{
		return nil, fmt.Errorf("Encountered error %q when retrieving tool matching %q from db", err, matcher)
	}
	return ret, nil
}


func GetTools(matcher interface{}) ([]*tool.Tool, error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(TOOLS)
	var ret []*tool.Tool
	err := tcol.Find(matcher).All(&ret)
	if err != nil{
		return nil, fmt.Errorf("Encountered error %q when retrieving tools matching %q from db", err, matcher)
	}
	return ret, nil
}


func AddFile(f *submission.File) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(FILES)
	err := col.Insert(f)
	if err != nil{
		return fmt.Errorf("Encountered error %q when adding file %q to db", err, f)
	}
	return nil
}

func AddSubmission(s *submission.Submission) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(SUBMISSIONS)
	err := col.Insert(s)
	if err != nil{
		return fmt.Errorf("Encountered error %q when adding submission %q to db", err, s)
	}
	return nil
}

func AddTool(t *tool.Tool) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(TOOLS)
	err := col.Insert(t)
	if err != nil{
		return fmt.Errorf("Encountered error %q when adding tool %q to db", err, t)
	}
	return nil
}

func AddResult(r *tool.Result) error{
	session := getSession()
	defer session.Close()
	col := session.DB(DB).C(RESULTS)
	err := col.Insert(r)
	if err != nil{
		return fmt.Errorf("Encountered error %q when adding result %q to db", err, r)
	}
	return nil
}


func Update(col string, matcher, change interface{}) error {
	session := getSession()
	defer session.Close()
	tcol := session.DB(DB).C(col)
	err := tcol.Update(matcher, change)
	if err != nil{
		return fmt.Errorf("Encountered error %q when updating %q matching %q to %q in db", err, col, matcher, change)
	}
	return nil
}


func AddUsers(users ...*user.User) error {
	session := getSession()
	defer session.Close()
	c := session.DB(DB).C(USERS)
	err := c.Insert(users)
	if err != nil{
		return fmt.Errorf("Encountered error %q when adding users %q to db", err, users)
	}
	return nil
}
