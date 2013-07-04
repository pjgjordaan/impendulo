package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
)

//File stores a single file's data from a submission.
type Test struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	Name      string        "name"
	User      string        "user"
	Package   string        "pkg"
	Time   int64         "time"
	Test      []byte        "test"
	Data      []byte        "data"
}

func (this *Test) Equals(that *Test) bool {
	return reflect.DeepEqual(this, that)
}

//NewFile
func NewTest(projectId bson.ObjectId, name, user, pkg string, test, data []byte) *Test {
	id := bson.NewObjectId()
	return &Test{id, projectId, name, user, pkg, util.CurMilis(), test, data}
}
