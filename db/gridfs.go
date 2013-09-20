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

package db

import (
	"encoding/gob"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	GRIDFS_NAME = "fs"
)

//HasGridFile checks whether this query needs to get data from GridFS
func HasGridFile(result tool.ToolResult, selector bson.M) bool {
	return (selector == nil || selector[project.DATA] == 1) && result.OnGridFS()
}

//GridFile loads a GridFile matching id into a provided data structure from GridFS.
func GridFile(id, ret interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	fs := session.DB("").GridFS(GRIDFS_NAME)
	var file *mgo.GridFile
	file, err = fs.OpenId(id)
	if err != nil {
		return
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	err = dec.Decode(ret)
	return
}

//AddGridFile creates a new GridFile and stores the provided data structure in it via gob.
func AddGridFile(id, data interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	fs := session.DB("").GridFS(GRIDFS_NAME)
	var file *mgo.GridFile
	file, err = fs.Create("")
	if err != nil {
		return
	}
	defer file.Close()
	file.SetId(id)
	enc := gob.NewEncoder(file)
	err = enc.Encode(data)
	return
}
