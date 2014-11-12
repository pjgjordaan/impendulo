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
	"github.com/godfried/impendulo/util/milliseconds"
	"labix.org/v2/mgo/bson"
)

type (
	//Submission is used for individual project submissions
	Submission struct {
		Id           bson.ObjectId `bson:"_id"`
		ProjectId    bson.ObjectId `bson:"projectid"`
		AssignmentId bson.ObjectId `bson:"assignmentid"`
		User         string        `bson:"user"`
		Mode         string        `bson:"mode"`
		Time         int64         `bson:"time"`
		Comments     []*Comment    `bson:"comments"`
	}
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
		"; Time: " + milliseconds.DateTimeString(s.Time)
}

//NewSubmission
func NewSubmission(pid, aid bson.ObjectId, u, m string, t int64) *Submission {
	return &Submission{bson.NewObjectId(), pid, aid, u, m, t, []*Comment{}}
}

func (s *Submission) Format(p *P) string {
	return fmt.Sprintf("%s \u2192 %s \u2192 %s", p.Name, s.User, milliseconds.DateTimeString(s.Time))
}

func (s *Submission) LoadComments() []*Comment {
	return s.Comments
}
