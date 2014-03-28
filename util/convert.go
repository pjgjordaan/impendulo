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
func ReadId(i interface{}) (bson.ObjectId, error) {
	switch v := i.(type) {
	case bson.ObjectId:
		return v, nil
	case string:
		if !bson.IsObjectIdHex(v) {
			return "", &CastError{"bson.ObjectId", v}
		} else {
			return bson.ObjectIdHex(v), nil
		}
	}
	return "", &CastError{"id", i}
}

//GetString converts a value in a map to a string.
func GetString(m map[string]interface{}, k string) (string, error) {
	i, ok := m[k]
	if !ok {
		return "", &MissingError{k}
	}
	switch v := i.(type) {
	case string:
		return v, nil
	}
	return fmt.Sprint(i), nil
}

//GetInt converts a value in a map to an int.
func GetInt(m map[string]interface{}, k string) (int, error) {
	i, ok := m[k]
	if !ok {
		return 0, &MissingError{k}
	}
	return Int(i)
}

//GetInt64 converts a value in a map to an int64.
func GetInt64(m map[string]interface{}, k string) (int64, error) {
	i, ok := m[k]
	if !ok {
		return 0, &MissingError{k}
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
func GetId(m map[string]interface{}, k string) (bson.ObjectId, error) {
	i, ok := m[k]
	if !ok {
		return "", &MissingError{k}
	}
	return ReadId(i)
}

//GetM converts a value in a map to a bson.M.
func GetM(m map[string]interface{}, k string) (bson.M, error) {
	i, ok := m[k]
	if !ok {
		return nil, &MissingError{k}
	}
	switch v := i.(type) {
	case bson.M:
		return v, nil
	case map[string]interface{}:
		return bson.M(v), nil
	}
	return nil, &CastError{"bson.M", i}
}

//GetBytes converts a value in a map to a []byte.
func GetBytes(m map[string]interface{}, k string) ([]byte, error) {
	i, ok := m[k]
	if !ok {
		return nil, &MissingError{k}
	}
	return toBytes(i)
}

//GetStrings converts a value in a map to a []string.
func GetStrings(m map[string]interface{}, k string) ([]string, error) {
	i, ok := m[k]
	if !ok {
		return nil, &MissingError{k}
	}
	return toStrings(i)
}

//toBytes converts an interface to a []byte.
func toBytes(i interface{}) ([]byte, error) {
	v, ok := i.([]byte)
	if !ok {
		return nil, &CastError{"[]byte", i}
	}
	return v, nil
}

//toStrings converts an interface to a []string.
func toStrings(i interface{}) ([]string, error) {
	v, ok := i.([]string)
	if !ok {
		a, ok := i.([]interface{})
		if !ok {
			return nil, &CastError{"[]string", i}
		}
		v = make([]string, len(a))
		for j, i := range a {
			s, ok := i.(string)
			if !ok {
				return nil, &CastError{"string", i}
			}
			v[j] = s
		}
	}
	return v, nil
}

//ToSet converts an array to a set.
func ToSet(vals []string) map[string]bool {
	s := make(map[string]bool)
	for _, v := range vals {
		s[v] = true
	}
	return s
}
