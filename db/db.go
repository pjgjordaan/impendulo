package db

import (
	"github.com/disco-volante/intlola/client"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
	"fmt"
)
const DB = "impendulo"
const USERS = "users"
const PROJECTS = "projects"
const ADDRESS = "localhost"
/*
Finds a user in the database. 
This is used to authenticate a login attempt.
*/
func ReadUser(uname string) (user *UserData, err error) {
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB).C(USERS)
		err = c.FindId(uname).One(&user)
	}else {
		panic(err)
	}
	return user, err
}

/*
The struct used to store user information in the database.
*/
type UserData struct{
	Name string "_id,omitempty"
	Password string "password"
}


func NewUser(uname, pword string) (*UserData){
	return &UserData{uname, pword}
}

/*
Adds or updates multiple users.
*/
func AddUsers(users...  *UserData)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB).C(USERS)
		for _, user := range users{
			_, err = c.UpsertId(user.Name, user)
			if err != nil{
				break
			}
		}		
	}else {
		panic(err)
	}
	return err
}


/*
A struct used to store information about individual project submissions
in the database.
*/
type Submission struct{
	Name string "name"
	User string "user"
	Date int64 "date"
	Subnum int "number"
	Format string "format"
	Files []FileData "files"
}

/*
A struct used to store individual files in the database.
*/
type FileData struct{
	Name string "name"
	Data [] byte "data"
	Date int64 "date"
}

/*
Creates a new project submission for a given user.
*/
func CreateSubmission(c *client.Client)(num int, err error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		pcol := session.DB(DB).C(PROJECTS)
		matcher := bson.M{"name" : c.Project, "user" : c.Name}
		num, err = pcol.Find(matcher).Count()
		if err == nil{
			date := time.Now().UnixNano()
			sub := &Submission{c.Project, c.Name, date, num, c.Format,  make([]FileData, 0, 300)}
			err = pcol.Insert(sub)
			var s *Submission
			pcol.Find(matcher).One(&s)
			fmt.Println(s, err)
		}
	} else{
		panic(err)
	}
	return num, err
}

/*
Adds a new file to a user's project submission.
*/
func AddFile(c *client.Client, fname string, data []byte)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		fcol := session.DB(DB).C(PROJECTS)
		date := time.Now().UnixNano()
		file := &FileData{fname, data, date}
		matcher := bson.M{"name" : c.Project, "user" : c.Name, "number" :c.SubNum}
		err = fcol.Update(matcher, bson.M{"$push": bson.M{ "files": file}})
	} else{
		panic(err)
	}
	return err
}

/*
Adds a new file to a user's project submission.
*/
func AddTests(project string, data []byte)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		fcol := session.DB(DB).C(PROJECTS)
		test := bson.M{"project" : project, "tests": data}
		_,err = fcol.Upsert(bson.M{"project" : project}, test)
	} else{
		panic(err)
	}
	return err
}


/*
Retrieves all distinct values for a given field.
*/
func GetAll(field string)(values []string, err error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		fcol := session.DB(DB).C(PROJECTS)
		err = fcol.Find(nil).Distinct(field, &values)
	} else{
		panic(err)
	}
	return values, err
}



