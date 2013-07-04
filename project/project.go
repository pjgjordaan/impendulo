package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"github.com/godfried/impendulo/util"
)

type Project struct {
	Id   bson.ObjectId "_id"
	Name string        "name"
	User string        "user"
	Lang string        "lang"
	Time int64         "time"
}

func (this *Project) String() string {
	return "Name: " + this.Name + "; User: " + this.User + "; Time: " + strconv.Itoa(int(this.Time))
}

func (this *Project) Equals(that *Project) bool {
	return reflect.DeepEqual(this, that)
}

func NewProject(name, user, lang string) *Project {
	id := bson.NewObjectId()
	return &Project{id, name, user, lang, util.CurMilis()}
}
