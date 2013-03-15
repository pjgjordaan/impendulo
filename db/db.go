package db

import (
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo"
        "labix.org/v2/mgo/bson"
)
const DB_NAME = "intlola"

func Read(collName, key, value string) (info [] *client.ClientData, err error) {
	session, err := mgo.Dial("localhost")
	defer session.Close()
	c := session.DB(DB_NAME).C(collName)
	err = c.Find(bson.M{key : value}).All(info)
	utils.Log("Read: ", info)
	return info, err
}

func AddOne(collName string, info *client.ClientData)(error){
	return Add(collName, []*client.ClientData{info})
}
func Add(collName string, infos[] *client.ClientData) (error) {
	session, err := mgo.Dial("localhost")
	defer session.Close()
	if err == nil{
		c := session.DB(DB_NAME).C(collName)
		err = c.Insert(infos)
		utils.Log("Inserted: ", infos)
	}
	return err
}

