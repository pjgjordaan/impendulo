package project

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"time"
)

//Submission is used for individual project submissions
type Submission struct {
	Id      bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	User    string        "user"
	Time    int64         "time"
	Mode    string        "mode"
}

//IsTest
func (s *Submission) IsTest() bool {
	return s.Mode == TEST_MODE
}

func (this *Submission) String() string {
	return "ProjectId: " + this.ProjectId.String() + "; User: " + this.User + "; Time: " + time.Unix(0, this.Time).String()

}

func (this *Submission) Date() string {
	return time.Unix(0, this.Time).String()
}

func (this *Submission) Equals(that *Submission) bool {
	return reflect.DeepEqual(this, that)
}

//NewSubmission
func NewSubmission(projectId bson.ObjectId, user, mode string) *Submission {
	subId := bson.NewObjectId()
	now := time.Now().UnixNano()
	return &Submission{subId, projectId, user, now, mode}
}

//isOutFolder
func isOutFolder(arg string) bool {
	return arg == SRC_DIR || arg == BIN_DIR
}