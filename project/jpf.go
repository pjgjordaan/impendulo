package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"strings"
)

type JPFFile struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	Name      string        "name"
	User      string        "user"
	Data      []byte        "data"
}

func (this *JPFFile) Equals(that *JPF) bool {
	return reflect.DeepEqual(this, that)
}

//NewFile
func NewJPFFile(projectId bson.ObjectId, name, user string, data []byte) *JPFFile {
	id := bson.NewObjectId()
	return &JPFFile{id, projectId, name, user, data}
}
