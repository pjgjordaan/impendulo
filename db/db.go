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
const TOOLS = "tools"
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
A struct used to store individual files in the database.
*/
type FileData struct {
	Id       bson.ObjectId "_id"
	SubId    bson.ObjectId "subid"
	Name     string        "name"
	FileType string        "type"
	Data     []byte        "data"
	Date     int64         "date"
	Results *bson.M "results"
}

func (f *FileData) IsSource() bool {
	return f.FileType == "SOURCE"
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
		file := &FileData{fileId, subId, fname, ftype, data, date, new(bson.M)}
		err = fcol.Insert(file)
	}
	return fileId, err
}

func AddResults(fileId bson.ObjectId, name, result string,  data []byte) (err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		fcol := session.DB(DB).C(FILES)
		matcher := bson.M{"_id": fileId}
		err = fcol.Update(matcher, bson.M{"$push": bson.M{"results": bson.M{name: bson.M{"result":result, "data": data}}}})
	}		
	return err
}

func GetById(id interface{}, col string) (ret interface{}, err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		c := session.DB(DB).C(col)
		err = c.FindId(id).One(&ret)
	}
	return ret, err
}

func GetAll(col string)(items []interface{}, err error){
	session, err := getSession()
	if err == nil {
		defer session.Close()
		tcol := session.DB(DB).C(col)
		err = tcol.Find(nil).All(&items)
	}
	return items, err	
}

func AddSingle(item interface{}, col string)(err error){
	session, err := getSession()
	if err == nil {
		defer session.Close()
		tcol := session.DB(DB).C(col)
		err = tcol.Insert(item)
	}
	return err	
}

func AddMany(col string, items... interface{})(err error) {
	session, err := getSession()
	if err == nil {
		defer session.Close()
		c := session.DB(DB).C(col)
		for _, item := range items {
			err = c.Insert(item)
			if err != nil {
				break
			}
		}
	}
	return err
}
