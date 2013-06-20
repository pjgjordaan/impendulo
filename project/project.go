package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"time"
)

type Project struct {
	Id   bson.ObjectId "_id"
	Name string        "name"
	User string        "user"
	Time int64         "time"
	Lang string        "lang"
}

func (this *Project) String() string {
	return "Name: " + this.Name + "; User: " + this.User + "; Time: " + time.Unix(0, this.Time).String()
}

func (this *Project) Date() string {
	return time.Unix(0, this.Time).String()
}

func (this *Project) Equals(that *Project) bool {
	return reflect.DeepEqual(this, that)
}

func NewProject(name, user, lang string) *Project {
	id := bson.NewObjectId()
	now := time.Now().UnixNano()
	return &Project{id, name, user, now, lang}
}
