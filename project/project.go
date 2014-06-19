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
		Id          bson.ObjectId `bson:"_id"`
		Name        string        `bson:"name"`
		User        string        `bson:"user"`
		Lang        string        `bson:"lang"`
		Time        int64         `bson:"time"`
		Description string        `bson:"description"`
	}

	Comment struct {
		User string
		Data []byte
	}
)

//TypeName
func (p *Project) TypeName() string {
	return "project"
}

//String
func (p *Project) String() string {
	return "Type: project.Project; Id: " + p.Id.Hex() +
		"; Name: " + p.Name + "; User: " + p.User +
		"; Lang: " + p.Lang + "; Time: " + util.Date(p.Time)
}

//New
func New(n, u, l, d string) *Project {
	return &Project{Id: bson.NewObjectId(), Name: n, User: u, Lang: l, Time: util.CurMilis(), Description: d}
}
