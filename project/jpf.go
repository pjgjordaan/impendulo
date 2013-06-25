package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"strings"
)

type JPF struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	Name      string        "name"
	User      string        "user"
	Package   string        "package"
	IsJava    bool          "isjava"
	Data      []byte        "data"
}

func (this *JPF) Equals(that *JPF) bool {
	return reflect.DeepEqual(this, that)
}

//NewFile
func NewJPF(projectId bson.ObjectId, name, user string, data []byte) *JPF {
	id := bson.NewObjectId()
	return &JPF{Id: id, ProjectId: projectId, Name: name, User: user, IsJava: strings.HasSuffix(name, "java"), Data: data}
}
