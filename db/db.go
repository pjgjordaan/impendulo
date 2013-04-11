package db

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

const DB = "impendulo"
const USERS = "users"
const PROJECTS = "projects"
const FILES = "files"
const ADDRESS = "localhost"

var activeSession *mgo.Session

func getSession() (s *mgo.Session, err error) {
	if activeSession == nil {
		activeSession, err = mgo.Dial(ADDRESS)
	}
	if err == nil {
		s = activeSession.Clone()
	}
	return s, err
}

/*
Finds a user in the database. 
This is used to authenticate a login attempt.
*/
func ReadUser(uname string) (user *UserData, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		c := session.DB(DB).C(USERS)
		err = c.FindId(uname).One(&user)
	}
	return user, err
}

/*
The struct used to store user information in the database.
*/
type UserData struct {
	Name     string "_id,omitempty"
	Password string "password"
	Salt     string "salt"
}

func NewUser(uname, pword, salt string) *UserData {
	return &UserData{uname, pword, salt}
}

/*
Adds or updates multiple users.
*/
func AddUsers(users ...*UserData) error {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		c := session.DB(DB).C(USERS)
		for _, user := range users {
			_, err = c.UpsertId(user.Name, user)
			if err != nil {
				break
			}
		}
	}
	return err
}

/*
A struct used to store information about individual project submissions
in the database.
*/
type Submission struct {
	Id      bson.ObjectId "_id"
	Project string        "project"
	User    string        "user"
	Date    int64         "date"
	Subnum  int           "number"
	Mode    string        "mode"
}

func (s *Submission) IsTest() bool {
	return s.Mode == "TEST"
}

/*
A struct used to store individual files in the database.
*/
type FileData struct {
	Id       bson.ObjectId "_id"
	SubId    bson.ObjectId "subid"
	Name     string        "name"
	FileType string        "type"
	Data     []byte        "data"
	Date     int64         "date"
}

func (f *FileData) IsSource() bool {
	return f.FileType == "SOURCE"
}

/*
Creates a new project submission for a given user.
*/
func CreateSubmission(project, user, mode string) (subId bson.ObjectId, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		pcol := session.DB(DB).C(PROJECTS)
		matcher := bson.M{"name": project, "user": user}
		num, err := pcol.Find(matcher).Count()
		if err == nil {
			date := time.Now().UnixNano()
			subId = bson.NewObjectId()
			sub := &Submission{subId, project, user, date, num, mode}
			err = pcol.Insert(sub)
		}
	}
	return subId, err
}

func GetSubmission(id bson.ObjectId) (sub *Submission, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		pcol := session.DB(DB).C(PROJECTS)
		matcher := bson.M{"_id": id}
		err = pcol.Find(matcher).One(&sub)
	}
	return sub, err
}

func GetTests(project string) (tests *FileData, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		pcol := session.DB(DB).C(PROJECTS)
		matcher := bson.M{"project": project, "mode": "TEST"}
		var sub *Submission
		err = pcol.Find(matcher).One(&sub)
		if err == nil {
			fcol := session.DB(DB).C(FILES)
			matcher = bson.M{"subid": sub.Id}
			err = fcol.Find(matcher).One(&tests)
		}
	}
	return tests, err
}

/*
Adds a new file to a user's project submission.
*/
func AddFile(subId bson.ObjectId, fname, ftype string, data []byte) (fileId bson.ObjectId, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		fcol := session.DB(DB).C(FILES)
		date := time.Now().UnixNano()
		fileId = bson.NewObjectId()
		file := &FileData{fileId, subId, fname, ftype, data, date}
		err = fcol.Insert(file)
	}
	return fileId, err
}

func GetFile(fileId bson.ObjectId) (f *FileData, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		fcol := session.DB(DB).C(FILES)
		matcher := bson.M{"_id": fileId}
		err = fcol.Find(matcher).One(&f)
	}
	return f, err
}

func AddResults(fileId bson.ObjectId, key string, data []byte) (err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		fcol := session.DB(DB).C(FILES)
		matcher := bson.M{"_id": fileId}
		err = fcol.Update(matcher, bson.M{"$push": bson.M{"results": bson.M{"type": key, "data": data}}})
	}
	return err
}

/*
Retrieves all distinct values for a given field.
*/
func GetAll(field string) (values []string, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		fcol := session.DB(DB).C(PROJECTS)
		err = fcol.Find(nil).Distinct(field, &values)
	}
	return values, err
}
