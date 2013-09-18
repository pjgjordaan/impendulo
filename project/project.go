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

//Package project provides data structures for storing information
//about projects, submissions and files.
package project

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Project represents a Impendulo project.
	Project struct {
		Id       bson.ObjectId "_id"
		Name     string        "name"
		User     string        "user"
		Lang     string        "lang"
		Time     int64         "time"
		Skeleton []byte        "skeleton"
	}
)

//TypeName
func (this *Project) TypeName() string {
	return "project"
}

//String
func (this *Project) String() string {
	return "Type: project.Project; Id: " + this.Id.Hex() +
		"; Name: " + this.Name + "; User: " + this.User +
		"; Lang: " + this.Lang + "; Time: " + util.Date(this.Time)
}

//NewProject
func NewProject(name, user, lang string, data []byte) *Project {
	id := bson.NewObjectId()
	return &Project{id, name, user, lang, util.CurMilis(), data}
}
