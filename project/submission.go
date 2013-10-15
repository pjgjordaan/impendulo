//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
		Status    Status        "status"
	}
	Status int
)

const (
	UNKNOWN Status = iota
	BUSY
	FAILED
	NOTJUNIT
	NOTJPF
	JUNIT_NOTJPF
	JPF_NOTJUNIT
	JUNIT
	JPF
	ALL
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

func (this *Submission) Result() string {
	switch this.Status {
	case UNKNOWN:
		return "Unknown status."
	case BUSY:
		return "Busy evaluating snapshots."
	case FAILED:
		return "Submission provided incorrect solution."
	case NOTJUNIT:
		return "Failed unit tests."
	case NOTJPF:
		return "Failed JPF evaluation."
	case JPF_NOTJUNIT:
		return "Passed JPF evaluation but not all unit tests passed."
	case JUNIT_NOTJPF:
		return "All unit tests passed, but failed JPF evaluation."
	case JUNIT:
		return "Successfuly passed unit tests."
	case JPF:
		return "Successfully passed JPF evaluation."
	case ALL:
		return "Successfully passed JPF and unit testing evaluation."
	default:
		return fmt.Sprintf("Invalid status %d.", this.Status)
	}
}

//NewSubmission
func NewSubmission(projectId bson.ObjectId, user, mode string, time int64) *Submission {
	subId := bson.NewObjectId()
	return &Submission{subId, projectId, user, mode, time, BUSY}
}
