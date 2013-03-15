package db

import (
	"github.com/disco-volante/intlola/client"
	"labix.org/v2/mgo"
        "labix.org/v2/mgo/bson"
)
const DB_NAME = "intlola"

func Read(collName, key, value string) (vals [] interface{}, err error) {
	session, err := mgo.Dial("localhost")
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(collName)
		err = c.Find(bson.M{key : value}).All(vals)
	}
	return vals, err
}

func ReadUser(uname, pword string) (user *client.ClientData, err error) {
	session, err := mgo.Dial("localhost")
	defer session.Close()
	c := session.DB(DB_NAME).C("users")
	err = c.Find(bson.M{"name" : uname, "password" : pword}).One(user)
	return user, err
}

func Add(collName string, vals [] interface{}) (error) {
	session, err := mgo.Dial("localhost")
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(collName)
		err = c.Insert(vals...)
	}
	return err
}

func AddUsers(users...  *client.ClientData)(error){
	session, err := mgo.Dial("localhost")
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C("users")
		for _, user := range users{ 
			_, err = c.Upsert(bson.M{"name" : user.Name}, user)
			if err != nil{
				break
			}
		}		
	}
	return err
}
