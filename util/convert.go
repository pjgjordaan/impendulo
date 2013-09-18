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

//GetInt64 converts a value in a map to an int.
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
