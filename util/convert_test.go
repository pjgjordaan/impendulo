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
	"errors"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
)

var testmap = map[string]interface{}{
	"string": "2a 3", "int": 2, "id": bson.NewObjectId(),
	"map": bson.M{"a": "b"}, "bytes": []byte("AB"),
	"strings": []string{"A", "B"}, "int64": int64(2231231223123123123),
}

func TestGetString(t *testing.T) {
	this := "string"
	for k, v := range testmap {
		res, err := GetString(testmap, k)
		if err != nil {
			t.Error(err)
		} else if k == this && res != v {
			t.Error(res, "!=", v)
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
		} else if err == nil && k != "int64" {
			t.Error(errors.New("Error function should not cast"))
		}
	}
}

func TestGetInt64(t *testing.T) {
	this := "int64"
	for k, v := range testmap {
		res, err := GetInt64(testmap, k)
		if k == this {
			if err != nil || res != v.(int64) {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil && k != "int" {
			t.Error(errors.New("Error function should not cast"))
		}
	}
}

func TestGetId(t *testing.T) {
	this := "id"
	for k, v := range testmap {
		res, err := GetId(testmap, k)
		if k == this {
			if err != nil || res != v {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil {
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
			if err != nil || !reflect.DeepEqual(res, val) {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil {
			t.Error(errors.New("Error function should not cast"))
		}
	}
}

func TestGetBytes(t *testing.T) {
	this := "bytes"
	for k, v := range testmap {
		res, err := GetBytes(testmap, k)
		if k == this {
			val, _ := toBytes(v)
			if err != nil || !bytes.Equal(res, val) {
				t.Error(err, res, "!=", v)
			}
		} else if err == nil {
			t.Error(errors.New("Error function should not cast"))
		}
	}
}

func TestGetStrings(t *testing.T) {
	this := "strings"
	for k, v := range testmap {
		res, err := GetStrings(testmap, k)
		if k == this {
			val, _ := toStrings(v)
			if err != nil || !reflect.DeepEqual(res, val) {
				t.Error(err)
			}
		} else if err == nil {
			t.Error(errors.New("Error function should not cast"), k, res)
		}
	}
}
