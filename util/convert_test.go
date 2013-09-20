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
