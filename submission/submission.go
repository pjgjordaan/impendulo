package submission

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"time"
)

//Submission is used for individual project submissions
type Submission struct {
	Id      bson.ObjectId "_id"
	Project string        "project"
	User    string        "user"
	Time    int64         "time"
	Mode    string        "mode"
	Lang    string        "lang"
}

//IsTest
func (s *Submission) IsTest() bool {
	return s.Mode == TEST_MODE
}

func (this *Submission) String() string {
	return "Project: " + this.Project + "; User: " + this.User + "; Time: " + time.Unix(0, this.Time).String()

}

func (this *Submission) Equals(that *Submission) bool {
	return reflect.DeepEqual(this, that)
}

//NewSubmission
func NewSubmission(project, user, mode, lang string) *Submission {
	subId := bson.NewObjectId()
	now := time.Now().UnixNano()
	return &Submission{subId, project, user, now, mode, lang}
}

//isOutFolder
func isOutFolder(arg string) bool {
	return arg == SRC_DIR || arg == BIN_DIR
}