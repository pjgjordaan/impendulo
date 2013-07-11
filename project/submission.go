package project

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"reflect"
)

//Submission is used for individual project submissions
type Submission struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	User      string        "user"
	Mode      string        "mode"
	Time      int64         "time"
}

//IsTest
func (this *Submission) IsTest() bool {
	return this.Mode == TEST_MODE
}

func (this *Submission) TypeName() string {
	return "submission"
}

func (this *Submission) String() string {
	return "Type: project.Submission; Id: " + this.Id.Hex() + "; ProjectId: " + this.ProjectId.Hex() + "; User: " + this.User + "; Mode: " + this.Mode + "; Time: " + util.Date(this.Time)
}

func (this *Submission) Equals(that *Submission) bool {
	return reflect.DeepEqual(this, that)
}

//NewSubmission
func NewSubmission(projectId bson.ObjectId, user, mode string, time int64) *Submission {
	subId := bson.NewObjectId()
	return &Submission{subId, projectId, user, mode, time}
}

//isOutFolder
func isOutFolder(arg string) bool {
	return arg == SRC_DIR || arg == BIN_DIR
}
