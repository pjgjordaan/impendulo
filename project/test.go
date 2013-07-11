package project

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"reflect"
)

//File stores a single file's data from a submission.
type Test struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	Name      string        "name"
	User      string        "user"
	Package   string        "pkg"
	Time      int64         "time"
	Test      []byte        "test"
	Data      []byte        "data"
}

func (this *Test) Equals(that *Test) bool {
	return reflect.DeepEqual(this, that)
}

func (this *Test) TypeName() string {
	return "test file"
}

func (this *Test) String() string {
	return "Type: project.Test; Id: " + this.Id.Hex() + "; ProjectId: " + this.ProjectId.Hex() + "; Name: " + this.Name + "; Package: " + this.Package + "; User: " + this.User + "; Time: " + util.Date(this.Time)
}

//NewFile
func NewTest(projectId bson.ObjectId, name, user, pkg string, test, data []byte) *Test {
	id := bson.NewObjectId()
	return &Test{id, projectId, name, user, pkg, util.CurMilis(), test, data}
}
