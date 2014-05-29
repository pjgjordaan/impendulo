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

package jpf

import (
	"bytes"
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"strings"
)

type (
	//Config represents a JPF configuration.
	//This means that it contains configured JPF properties specific to running a certain project.
	Config struct {
		Id        bson.ObjectId `bson:"_id"`
		ProjectId bson.ObjectId `bson:"projectid"`
		Time      int64         `bson:"time"`
		Target    *tool.Target  `bson:"target"`
		//Contains configured JPF properties
		Data []byte `bson:"data"`
	}
	empty struct{}
)

var (
	//reserved are JPF properties which are not allowed to be set by the user.
	reserved = map[string]empty{
		"search.class":                  empty{},
		"listener":                      empty{},
		"target":                        empty{},
		"report.publisher":              empty{},
		"report.xml.class":              empty{},
		"report.xml.file":               empty{},
		"classpath":                     empty{},
		"report.xml.start":              empty{},
		"report.xml.transition":         empty{},
		"report.xml.constraint":         empty{},
		"report.xml.property_violation": empty{},
		"report.xml.show_steps":         empty{},
		"report.xml.show_method":        empty{},
		"report.xml.show_code":          empty{},
		"report.xml.finished":           empty{},
	}
)

//Allowed checks whether a JPF property is allowed to be configured.
func Allowed(key string) bool {
	_, ok := reserved[key]
	return !ok
}

//String
func (this *Config) String() string {
	return "Type: project.Config; Id: " + this.Id.Hex() +
		"; ProjectId: " + this.ProjectId.Hex() +
		"; Time: " + util.Date(this.Time)
}

//NewConfig creates a new JPF configuration for a certain project.
func NewConfig(projectId bson.ObjectId, target *tool.Target, data []byte) *Config {
	return &Config{
		Id:        bson.NewObjectId(),
		ProjectId: projectId,
		Time:      util.CurMilis(),
		Target:    target,
		Data:      data,
	}
}

//JPFBytes converts a map of JPF properties to a byte array which can be used when running JPF.
func JPFBytes(config map[string][]string) (ret []byte, err error) {
	buff := new(bytes.Buffer)
	for key, val := range config {
		_, err = buff.WriteString(fmt.Sprintf("%s = %s\n", key, strings.Join(val, ",")))
		if err != nil {
			return
		}
	}
	ret = buff.Bytes()
	return
}
