//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
