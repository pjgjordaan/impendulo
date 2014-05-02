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
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/godfried/impendulo/util/errors"

	"io"

	"labix.org/v2/mgo/bson"

	"os"
)

//ReadJSON reads all JSON data from a reader.
func ReadJSON(r io.Reader) (map[string]interface{}, error) {
	b, e := ReadData(r)
	if e != nil {
		return nil, e
	}
	var m map[string]interface{}
	if e = json.Unmarshal(b, &m); e != nil {
		return nil, errors.NewUtil(b, "unmarshalling data from", e)
	}
	return m, nil
}

//LoadMap loads a map stored in a file.
func LoadMap(n string) (map[bson.ObjectId]bool, error) {
	f, e := os.Open(n)
	if e != nil {
		return nil, errors.NewUtil(n, "opening", e)
	}
	d := gob.NewDecoder(f)
	var m map[bson.ObjectId]bool
	if e = d.Decode(&m); e != nil {
		return nil, errors.NewUtil(n, "decoding map stored in", e)
	}
	return m, nil
}

//SaveMap saves a map to the filesystem.
func SaveMap(m map[bson.ObjectId]bool, n string) error {
	f, e := os.Create(n)
	if e != nil {
		return errors.NewUtil(n, "creating", e)
	}
	enc := gob.NewEncoder(f)
	if e = enc.Encode(&m); e != nil {
		return errors.NewUtil(m, "encoding map", e)
	}
	return nil
}

//WriteJSON writes JSON marshalled data to to the writer.
func WriteJSON(w io.Writer, i interface{}) error {
	m, e := json.Marshal(i)
	if e != nil {
		return errors.NewUtil(i, "marshalling json", e)
	}
	if _, e = w.Write(m); e != nil {
		return errors.NewUtil(m, "writing json", e)
	}
	return nil
}

func JSON(i interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	if e := WriteJSON(b, i); e != nil {
		return nil, e
	}
	return b.Bytes(), nil
}
