package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"labix.org/v2/mgo/bson"
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

func TestMerge(t *testing.T) {
	id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	id3 := bson.NewObjectId()
	id4 := bson.NewObjectId()
	id5 := bson.NewObjectId()
	m1 := map[bson.ObjectId]bool{id1: true, id2: false, id3: false, id5: true}
	m2 := map[bson.ObjectId]bool{id1: true, id2: true, id4: true, id5: false}
	Merge(m1, m2)
	if len(m1) != 5 {
		t.Error(errors.New("Elements not copied"))
	}
	for k, v := range m1 {
		if v2, ok := m2[k]; ok {
			if !v == v2 {
				t.Error(errors.New("Elements not merged correctly"))

			}
		}
	}
}

func TestReadBytes(t *testing.T) {
	orig := []byte("bytes")
	buff := bytes.NewBuffer(orig)
	ret := readBytes(buff)
	if !bytes.Equal(orig, ret) {
		t.Error(errors.New("Bytes not equal"))
	}
}

func TestZip(t *testing.T) {
	files := map[string][]byte{"readme.txt": []byte("This archive contains some text files."), "gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"), "todo.txt": []byte("Get animal handling licence.\nWrite more examples.")}
	zipped, err := Zip(files)
	if err != nil {
		t.Error(err)
	}
	unzipped, err := UnzipToMap(zipped)
	if err != nil {
		t.Error(err)
	}
	if len(files) != len(unzipped) {
		t.Error(errors.New("Zip error; invalid size"))
	}
	for k, v := range files {
		if !bytes.Equal(v, unzipped[k]) {
			t.Error(errors.New("Zip error, values not equal."))
		}
	}

}
