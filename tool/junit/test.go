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

package junit

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Test stores tests for a project.
	Test struct {
		Id        bson.ObjectId "_id"
		ProjectId bson.ObjectId "projectid"
		Name      string        "name"
		User      string        "user"
		Package   string        "pkg"
		Time      int64         "time"
		//The test file
		Test []byte "test"
		//The data files needed for the test stored in a zip archive
		Data []byte "data"
	}
)

//String
func (this *Test) String() string {
	return "Type: junit.Test; Id: " + this.Id.Hex() +
		"; ProjectId: " + this.ProjectId.Hex() +
		"; Name: " + this.Name + "; Package: " + this.Package +
		"; User: " + this.User + "; Time: " + util.Date(this.Time)
}

//NewTest
func NewTest(projectId bson.ObjectId, name, user, pkg string, test, data []byte) *Test {
	id := bson.NewObjectId()
	return &Test{id, projectId, name, user, pkg, util.CurMilis(), test, data}
}
