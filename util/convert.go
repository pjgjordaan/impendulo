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
	"fmt"
	"labix.org/v2/mgo/bson"
	"strconv"
)

//ReadId tries to read a bson.ObjectId from a string.
func ReadId(idStr string) (id bson.ObjectId, err error) {
	if !bson.IsObjectIdHex(idStr) {
		err = &CastError{"bson.ObjectId", idStr}
	} else {
		id = bson.ObjectIdHex(idStr)
	}
	return
}

//GetString converts a value in a map to a string.
func GetString(jobj map[string]interface{}, key string) (val string, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch v := ival.(type) {
	case string:
		val = v
	default:
		val = fmt.Sprint(v)
	}
	return
}

//GetInt converts a value in a map to an int.
func GetInt(jobj map[string]interface{}, key string) (val int, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch v := ival.(type) {
	case int64:
		val = int(v)
	case int:
		val = v
	case float64:
		val = int(v)
	case string:
		val, err = strconv.Atoi(v)
	default:
		err = &CastError{"int", v}
	}
	return
}

//GetInt64 converts a value in a map to an int64.
func GetInt64(jobj map[string]interface{}, key string) (val int64, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch v := ival.(type) {
	case int64:
		val = v
	case int:
		val = int64(v)
	case float64:
		val = int64(v)
	case string:
		val, err = strconv.ParseInt(v, 10, 64)
	default:
		err = &CastError{"int", v}
	}
	return
}

//GetId converts a value in a map to a bson.ObjectId.
func GetId(jobj map[string]interface{}, key string) (id bson.ObjectId, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch v := ival.(type) {
	case bson.ObjectId:
		id = v
	case string:
		id, err = ReadId(v)
	default:
		err = &CastError{"id", v}
	}
	return
}

//GetM converts a value in a map to a bson.M.
func GetM(jobj map[string]interface{}, key string) (val bson.M, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch v := ival.(type) {
	case bson.M:
		val = v
	default:
		err = &CastError{"bson.M", v}
	}
	return
}

//GetBytes converts a value in a map to a []byte.
func GetBytes(jobj map[string]interface{}, key string) ([]byte, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, &MissingError{key}
	}
	return toBytes(ival)
}

//GetStrings converts a value in a map to a []string.
func GetStrings(jobj map[string]interface{}, key string) ([]string, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, &MissingError{key}
	}
	return toStrings(ival)
}

//toBytes converts an interface to a []byte.
func toBytes(ival interface{}) ([]byte, error) {
	val, ok := ival.([]byte)
	if !ok {
		return nil, &CastError{"[]byte", ival}
	}
	return val, nil
}

//toStrings converts an interface to a []string.
func toStrings(ivals interface{}) ([]string, error) {
	vals, ok := ivals.([]string)
	if !ok {
		islice, ok := ivals.([]interface{})
		if !ok {
			return nil, &CastError{"[]string", ivals}
		}
		vals = make([]string, len(islice))
		for i, ival := range islice {
			val, ok := ival.(string)
			if !ok {
				return nil, &CastError{"string", ival}
			}
			vals[i] = val
		}
	}
	return vals, nil
}

//ToSet converts an array to a set.
func ToSet(vals []string) map[string]bool {
	ret := make(map[string]bool)
	for _, v := range vals {
		ret[v] = true
	}
	return ret
}
