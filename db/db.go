package db

import (
	"github.com/disco-volante/intlola/client"
	"labix.org/v2/mgo"
)
const DB_NAME = "intlola"
const USERS = "users"
const FILES = "files"
const ADDRESS = "localhost"

func ReadUser(uname string) (user *UserData, err error) {
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(USERS)
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
		c := session.DB(DB_NAME).C(USERS)
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

type FileData struct{
	Name string "name"
	Project string "project"
	User string "user"
	Token string "token"
	Data []byte "data"
}

func AddFile(c *client.Client, fname string, data []byte)(error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		fcol := session.DB(DB_NAME).C(FILES)
		holder := &FileData{fname, c.Project, c.Name, c.Token, data}
		err = fcol.Insert(holder)	
	} else{
		panic(err)
	}
	return err
}

func GetTokens()(tokens []string, err error){
	session, err := mgo.Dial(ADDRESS)
	defer session.Close()
	if err == nil{
		fcol := session.DB(DB_NAME).C(FILES)
		err = fcol.Find(nil).Distinct("token", &tokens)
	} else{
		panic(err)
	}
	return tokens, err
}


