package project

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type Project struct {
	Id       bson.ObjectId "_id"
	Name     string        "name"
	User     string        "user"
	Lang     string        "lang"
	Time     int64         "time"
	Skeleton []byte        "skeleton"
}

func (this *Project) TypeName() string {
	return "project"
}

func (this *Project) String() string {
	return "Type: project.Project; Id: " + this.Id.Hex() + "; Name: " + this.Name + "; User: " + this.User + "; Lang: " + this.Lang + "; Time: " + util.Date(this.Time)
}

func NewProject(name, user, lang string, data []byte) *Project {
	id := bson.NewObjectId()
	return &Project{id, name, user, lang, util.CurMilis(), data}
}
