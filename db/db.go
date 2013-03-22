package db

import (
	"github.com/disco-volante/intlola/client"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)
const DB = "impendulo"
const USERS = "users"
const PROJECTS = "projects"
const ADDRESS = "localhost"

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


type UserData struct{
	Name string "_id,omitempty"
	Password string "password"
}
func NewUser(uname, pword string) (*UserData){
	return &UserData{uname, pword}
}
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

type ProjectData struct{
	Name string "name"
	User string "user"
	Date int64 "date"
	Token string "token"
	files []FileData "files"
}

type FileData struct{
	Name string "name"
	Data [] byte "data"
	Date int64 "date"
}

func CreateProject(c *client.Client)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		pcol := session.DB(DB).C(PROJECTS)
		date := time.Now().UnixNano()
		holder := &ProjectData{c.Project, c.Name, date,c.Token, make([]FileData, 0, 300)}
		err = pcol.Insert(holder)	
	} else{
		panic(err)
	}
	return err
}


func AddFile(c *client.Client, fname string, data []byte)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		fcol := session.DB(DB).C(PROJECTS)
		date := time.Now().UnixNano()
		file := &FileData{fname, data, date}
		matcher := bson.M{"name" : c.Project, "user" : c.Name, "token" : c.Token}
		err = fcol.Update(matcher, bson.M{"$push": bson.M{ "files": file}})
	} else{
		panic(err)
	}
	return err
}

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



