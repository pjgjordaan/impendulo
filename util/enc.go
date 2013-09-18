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
