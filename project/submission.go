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
		Id        bson.ObjectId `bson:"_id"`
		ProjectId bson.ObjectId `bson:"projectid"`
		User      string        `bson:"user"`
		Mode      string        `bson:"mode"`
		Time      int64         `bson:"time"`
		Status    Status        `bson:"status"`
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
func (s *Submission) SetMode(m string) error {
	if m != FILE_MODE && m != ARCHIVE_MODE {
		return fmt.Errorf("Unknown mode %s.", m)
	}
	s.Mode = m
	return nil
}

//String
func (s *Submission) String() string {
	return "Type: project.Submission; Id: " + s.Id.Hex() +
		"; ProjectId: " + s.ProjectId.Hex() +
		"; User: " + s.User + "; Mode: " + s.Mode +
		"; Time: " + util.Date(s.Time)
}

func (s *Submission) Result() string {
	switch s.Status {
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
		return fmt.Sprintf("Invalid status %d.", s.Status)
	}
}

//NewSubmission
func NewSubmission(pid bson.ObjectId, u, m string, t int64) *Submission {
	return &Submission{bson.NewObjectId(), pid, u, m, t, BUSY}
}

func (s *Submission) Format(p *Project) string {
	return fmt.Sprintf("%s \u2192 %s \u2192 %s", p.Name, s.User, util.Date(s.Time))
}
