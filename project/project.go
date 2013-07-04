package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
)

type Project struct {
	Id   bson.ObjectId "_id"
	Name string        "name"
	User string        "user"
	Lang string        "lang"
	Time int64         "time"
}

func (this *Project) TypeName() string{
	return "project"
}

func (this *Project) String() string {
	return "Type: project.Project; Id: "+this.Id.Hex()+"; Name: " + this.Name + "; User: " + this.User + "; Lang: " + this.Lang + "; Time: "+ util.Date(this.Time)
}

func (this *Project) Equals(that *Project) bool {
	return reflect.DeepEqual(this, that)
}

func NewProject(name, user, lang string) *Project {
	id := bson.NewObjectId()
	return &Project{id, name, user, lang, util.CurMilis()}
}
