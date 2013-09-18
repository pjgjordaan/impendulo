package project

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Submission is used for individual project submissions
	Submission struct {
		Id        bson.ObjectId "_id"
		ProjectId bson.ObjectId "projectid"
		User      string        "user"
		Mode      string        "mode"
		Time      int64         "time"
	}
)

//SetMode
func (this *Submission) SetMode(mode string) error {
	if mode != FILE_MODE && mode != ARCHIVE_MODE {
		return fmt.Errorf("Unknown mode %s.", mode)
	}
	this.Mode = mode
	return nil
}

//String
func (this *Submission) String() string {
	return "Type: project.Submission; Id: " + this.Id.Hex() +
		"; ProjectId: " + this.ProjectId.Hex() +
		"; User: " + this.User + "; Mode: " + this.Mode +
		"; Time: " + util.Date(this.Time)
}

//NewSubmission
func NewSubmission(projectId bson.ObjectId, user, mode string, time int64) *Submission {
	subId := bson.NewObjectId()
	return &Submission{subId, projectId, user, mode, time}
}
