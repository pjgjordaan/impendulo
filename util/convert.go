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
func ReadId(ival interface{}) (id bson.ObjectId, err error) {
	switch v := ival.(type) {
	case bson.ObjectId:
		id = v
	case string:
		if !bson.IsObjectIdHex(v) {
			err = &CastError{"bson.ObjectId", v}
		} else {
			id = bson.ObjectIdHex(v)
		}
	default:
		err = &CastError{"id", v}
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
func GetInt(m map[string]interface{}, key string) (int, error) {
	i, ok := m[key]
	if !ok {
		return 0, &MissingError{key}
	}
	return Int(i)
}

//GetInt64 converts a value in a map to an int64.
func GetInt64(m map[string]interface{}, key string) (int64, error) {
	i, ok := m[key]
	if !ok {
		return 0, &MissingError{key}
	}
	return Int64(i)
}

func Int(i interface{}) (int, error) {
	switch v := i.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	}
	return 0, &CastError{"int", i}
}

func Int64(i interface{}) (int64, error) {
	switch v := i.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	}
	return 0, &CastError{"int64", i}
}

func Float64(i interface{}) (float64, error) {
	switch v := i.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	}
	return 0.0, &CastError{"float64", i}
}

//GetId converts a value in a map to a bson.ObjectId.
func GetId(jobj map[string]interface{}, key string) (id bson.ObjectId, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	id, err = ReadId(ival)
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
