package utils
import (
	"errors"
	"labix.org/v2/mgo/bson"
	"testing"
	"bytes"
)

var testmap = map[string]interface{}{"string": "2 3", "int": 2, "id": bson.NewObjectId(), "map": bson.M{"a":"b"}, "bytes": []byte("AB"), "strings":  []string{"A","B"}}

func TestGetString(t *testing.T) {
	this := "string"
	for k, v := range testmap {
		res, err := GetString(testmap, k)
		if k == this {
			if err != nil || res != v {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil{
			t.Error(errors.New("Error function should not cast"))
		}
	}
}

func TestGetInt(t *testing.T) {
	this := "int"
	for k, v := range testmap {
		res, err := GetInt(testmap, k)
		if k == this {
			if err != nil || res != v {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil{
			t.Error(errors.New("Error function should not cast"))
		}
	}
}


func TestGetID(t *testing.T) {
	this := "id"
	for k, v := range testmap {
		res, err := GetID(testmap, k)
		if k == this {
			if err != nil || res != v {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil{
			t.Error(errors.New("Error function should not cast"))
		}
	}
}


func TestGetMap(t *testing.T) {
	this := "map"
	for k, v := range testmap {
		res, err := GetM(testmap, k)
		if k == this {
			val := v.(bson.M)
			if err != nil || !MEqual(res, val) {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil{
			t.Error(errors.New("Error function should not cast"))
		}
	}
}


func TestGetBytes(t *testing.T) {
	this := "bytes"
	for k, v := range testmap {
		res, err := GetBytes(testmap, k)
		if k == this {
			val, _ := ToBytes(v)
			if err != nil || !bytes.Equal(res, val) {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil{
			t.Error(errors.New("Error function should not cast"))
		}
	}
}


func TestGetStrings(t *testing.T) {
	this := "strings"
	for k, v := range testmap {
		res, err := GetStrings(testmap, k)
		if k == this {
			val,_ := ToStrings(v)
			if err != nil || !StringsEqual(res, val) {
				t.Error(err)
			}
		} else if err == nil{
			t.Error(errors.New("Error function should not cast"), k, res)
		}
	}
}
