package submission

import (
	"labix.org/v2/mgo/bson"
	"reflect"
)

//File stores a single file's data from a submission. 
type Test struct {
	Id      bson.ObjectId "_id"
	Project string "project"
	Package string "pkg"
	Lang string "lang"
	Names []string "names"
	Tests    []byte        "tests"
	Data  []byte  "data"
}

func (this *Test) Equals(that *Test) bool {
	return reflect.DeepEqual(this, that)
}

//NewFile
func NewTest(project, pkg, lang string, names []string, tests, data []byte) *Test {
	id := bson.NewObjectId()
	return &Test{id, project, pkg, lang, names, tests, data}
}
