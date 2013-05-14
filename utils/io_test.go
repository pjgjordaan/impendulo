package utils

import (
	"bytes"
	"encoding/json"
	"testing"
	"labix.org/v2/mgo/bson"
	"errors"
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

func TestGetString(t *testing.T) {
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D": "''"}
	for k, v := range jmap {
		if res, err := GetString(jmap, k); err != nil || res != v {
			t.Error(err, res, "!=", v)
		}
	}

}

func TestMerge(t *testing.T){
	id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	id3 := bson.NewObjectId()
	id4 := bson.NewObjectId()
	id5 := bson.NewObjectId()
	m1 := map[bson.ObjectId]bool{id1 : true, id2 : false, id3 : false, id5 : true}
	m2 := map[bson.ObjectId]bool{id1 : true, id2 : true, id4 : true, id5 : false}
	Merge(m1, m2)
	if len(m1) != 5{
		t.Error(errors.New("Elements not copied"))
	}
	for k, v := range m1{
		if v2, ok := m2[k]; ok{
			if !v == v2{
				t.Error(errors.New("Elements not merged correctly"))

			}
		}
	}
}
