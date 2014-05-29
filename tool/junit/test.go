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

package junit

import (
	"bytes"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Test stores tests for a project.
	Test struct {
		Id        bson.ObjectId `bson:"_id"`
		ProjectId bson.ObjectId `bson:"projectid"`
		Name      string        `bson:"name"`
		Package   string        `bson:"pkg"`
		Time      int64         `bson:"time"`
		Type      Type          `bson:"type"`
		Target    *tool.Target  `bson:"target"`
		//The test file
		Test []byte `bson:"test"`
		//The data files needed for the test stored in a zip archive
		Data []byte `bson:"data"`
	}

	Type     int
	TestType struct {
		ID   Type
		Name string
	}
)

const (
	DEFAULT Type = iota
	ADMIN
	USER
)

func TestTypes() []TestType {
	return []TestType{{DEFAULT, DEFAULT.String()}, {USER, USER.String()}, {ADMIN, ADMIN.String()}}
}

func (t Type) String() string {
	switch t {
	case DEFAULT:
		return "Default"
	case ADMIN:
		return "Admin"
	case USER:
		return "User"
	default:
		return "Unknown"
	}
}

//NewTest
func NewTest(projectId bson.ObjectId, name string, tipe Type, target *tool.Target, test, data []byte) *Test {
	return &Test{
		Id:        bson.NewObjectId(),
		ProjectId: projectId,
		Name:      name,
		Package:   util.GetPackage(bytes.NewReader(test)),
		Time:      util.CurMilis(),
		Type:      tipe,
		Target:    target,
		Test:      test,
		Data:      data,
	}
}
