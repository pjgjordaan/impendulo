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
	"bytes"
	"encoding/json"
	"errors"
	"labix.org/v2/mgo/bson"
	"os"
	"testing"
)

func TestReadJson(t *testing.T) {
	//This only
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D": "''"}
	marshalled, err := json.Marshal(jmap)
	if err != nil {
		t.Error(err)
	}
	reader := bytes.NewBuffer(marshalled)
	res, err := ReadJson(reader)
	if err != nil {
		t.Error(err)
	}
	for k, v := range jmap {
		if res[k] != v {
			t.Error(res[k], " != ", v)
		}
	}
}

func TestMapStorage(t *testing.T) {
	id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	id3 := bson.NewObjectId()
	id4 := bson.NewObjectId()
	m1 := map[bson.ObjectId]bool{id1: true, id2: false, id3: false, id4: true}
	gobFile := "test.gob"
	err := SaveMap(m1, gobFile)
	if err != nil {
		t.Error(err, "Error saving map")
	}
	defer os.Remove(gobFile)
	m2, err := LoadMap(gobFile)
	if err != nil {
		t.Error(err, "Error loading map")
	}
	if len(m1) != len(m2) {
		t.Error(errors.New("Error loading map; invalid size"))
	}
	for k, v := range m1 {
		if v != m2[k] {
			t.Error(errors.New("Error loading map, values not equal."))
		}
	}
}
