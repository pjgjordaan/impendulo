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
	"encoding/json"
	"errors"

	ue "github.com/godfried/impendulo/util/errors"
	"labix.org/v2/mgo/bson"

	"os"
	"testing"
)

func TestReadJSON(t *testing.T) {
	//This only
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D": "''"}
	marshalled, err := json.Marshal(jmap)
	if err != nil {
		t.Error(err)
	}
	reader := bytes.NewBuffer(marshalled)
	res, err := ReadJSON(reader)
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

func TestWriteJSON(t *testing.T) {
	m := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D": "''"}
	w := new(bytes.Buffer)
	if e := WriteJSON(w, m); e != nil {
		t.Error(e)
	}
	r, e := ReadJSON(w)
	if e != nil {
		t.Error(e)
	}
	for k, v := range m {
		if r[k] != v {
			t.Error(r[k], " != ", v)
		}
	}
	if e = WriteJSON(new(ue.Writer), m); e == nil {
		t.Error(errors.New("Expected error for error writer."))
	}
	m = map[string]interface{}{
		"B": func(msg string) bool { return msg != "" },
		"A": make(chan bool),
	}
	w.Reset()
	if e = WriteJSON(w, m); e == nil {
		t.Error(errors.New("Expected error for bad Json data."))
	}

}
