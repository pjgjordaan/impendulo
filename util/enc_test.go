package util

import (
	"bytes"
	"encoding/json"
	"testing"
	"labix.org/v2/mgo/bson"
	"errors"
)


func TestReadJson(t *testing.T) {
	//This only
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D":"''"}
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
	err := SaveMap(m1, "test.gob")
	if err != nil {
		t.Error(err, "Error saving map")
	}
	m2, err := LoadMap("test.gob")
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