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
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"
)

const (
	GRIDFS_NAME = "fs"
)

//HasGridFile checks whether this query needs to get data from GridFS
func HasGridFile(r result.Tooler, sl bson.M) bool {
	return (sl == nil || sl[REPORT] == 1) && r.OnGridFS()
}

//GridFile loads a GridFile matching id into a provided data structure from GridFS.
func GridFile(id, ret interface{}) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	f, e := s.DB("").GridFS(GRIDFS_NAME).OpenId(id)
	if e != nil {
		return e
	}
	defer f.Close()
	return gob.NewDecoder(f).Decode(ret)
}

//AddGridFile creates a new GridFile and stores the provided data structure in it via gob.
func AddGridFile(id, data interface{}) error {
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	f, e := s.DB("").GridFS(GRIDFS_NAME).Create("")
	if e != nil {
		return e
	}
	defer f.Close()
	f.SetId(id)
	return gob.NewEncoder(f).Encode(data)
}
