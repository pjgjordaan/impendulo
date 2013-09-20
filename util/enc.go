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

package util

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"labix.org/v2/mgo/bson"
	"os"
)

//ReadJson reads all Json data from a reader.
func ReadJson(r io.Reader) (jmap map[string]interface{}, err error) {
	read, err := ReadData(r)
	if err != nil {
		return
	}
	var holder interface{}
	err = json.Unmarshal(read, &holder)
	if err != nil {
		err = &UtilError{read, "unmarshalling data from", err}
		return
	}
	jmap, ok := holder.(map[string]interface{})
	if !ok {
		err = &UtilError{holder, "casting to json", nil}
	}
	return
}

//LoadMap loads a map stored in a file.
func LoadMap(fname string) (ret map[bson.ObjectId]bool, err error) {
	f, err := os.Open(fname)
	if err != nil {
		err = &UtilError{fname, "opening", err}
		return
	}
	dec := gob.NewDecoder(f)
	err = dec.Decode(&ret)
	if err != nil {
		err = &UtilError{fname, "decoding map stored in", err}
	}
	return
}

//SaveMap saves a map to the filesystem.
func SaveMap(mp map[bson.ObjectId]bool, fname string) (err error) {
	f, err := os.Create(fname)
	if err != nil {
		err = &UtilError{fname, "creating", err}
		return
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(&mp)
	if err != nil {
		err = &UtilError{mp, "encoding map", err}
	}
	return
}

//WriteJson writes json marshalled data to to the writer.
func WriteJson(w io.Writer, data interface{}) (err error) {
	marshalled, err := json.Marshal(data)
	if err != nil {
		err = &UtilError{data, "marshalling json", err}
		return
	}
	_, err = w.Write(marshalled)
	if err != nil {
		err = &UtilError{marshalled, "writing json", err}
	}
	return
}
