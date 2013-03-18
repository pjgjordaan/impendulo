package db

import (
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo"
        "labix.org/v2/mgo/bson"
)
const DB_NAME = "intlola"
const USERS = "users"
const PROJECTS = "projects"
const ADDRESS = "localhost"
func Read(collName, key, value string) (vals [] interface{}, err error) {
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(collName)
		err = c.Find(bson.M{key : value}).All(vals)
	} else {
		panic(err)
	}
	return vals, err
}

func ReadUser(uname, pword string) (user *client.ClientData, err error) {
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(USERS)
		var data *UserData
		holder := bson.M{"name" : uname, "password" : pword}
		err = c.Find(holder).One(&data)
		user = &client.ClientData{data.Name, data.Password,data.Projects}
	}else {
		panic(err)
	}
	return user, err
}

func Add(collName string, vals [] interface{}) (error) {
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(collName)
		err = c.Insert(vals...)
	}else {
		panic(err)
	}
	return err
}
type UserData struct{
	_Id interface{}
	Name string
	Password string
	Projects []string
}
func AddUsers(users...  *client.ClientData)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(USERS)
		for _, user := range users{
			holder :=  bson.M{"name": user.Name, "password" : user.Password, "projects" : make([]string,0,100)}
			_, err = c.Upsert(bson.M{"name" : user.Name}, holder)
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
	_Id interface{}
	Name string
	Number int
	User string
	Files []FileData
}

type FileData struct{
	Name string
	Data []byte
}

func AddFile(c *client.Client, fname string, data []byte)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		users := session.DB(DB_NAME).C(USERS)
		var user *UserData
			err := users.Find(bson.M{"name":c.Name}).One(&user)
		projects := session.DB(DB_NAME).C(PROJECTS)
		matcher := bson.M{"name":c.Project, "number":c.ProjectNum, "user" :user._Id} 
		q := projects.Find(matcher)
		if n, err := q.Count();err == nil && n == 0{
			holder := bson.M{"name" : c.Project, "number" : c.ProjectNum, "user" : user._Id, "files" : make([] FileData, 0, 100)}
			err = projects.Insert(holder)
		}
		if err == nil{
			projects.Update(matcher,  bson.M{"$push" : bson.M{"files": bson.M{"name" : fname, "data": data}}})
			var project *ProjectData
			err = projects.Find(matcher).One(&project)
			users.Update(user, bson.M{"$push" : bson.M{"projects":project._Id}})	
		}	
	} else{
		panic(err)
	}
	return err
}


func CreateProject(uname, pname string)(n int, err error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		users := session.DB(DB_NAME).C(USERS)
		var user *UserData
		if err = users.Find(bson.M{"name":uname}).One(&user); err == nil{
			projects := session.DB(DB_NAME).C(PROJECTS)
			matcher := bson.M{"name":pname, "user" :user._Id} 
			utils.Log(user._Id)
			q := projects.Find(matcher)
			if n, err = q.Count();err == nil {
				holder := bson.M{"name" : pname, "number" :n+1, "user" : user._Id, "files" : make([] FileData, 0, 100)}
				err = projects.Insert(holder)
				var project *ProjectData
				matcher = bson.M{"name":pname, "number":n+1, "user" :user._Id}
				if err = projects.Find(matcher).One(&project); err == nil{
						users.Update(user, bson.M{"$push" : bson.M{"projects":project._Id}})	
				}
			}
		}	
	} else{
		panic(err)
	}
	return n, err
}