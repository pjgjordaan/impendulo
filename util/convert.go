package util

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"strconv"
)

type MissingError struct {
	key string
}

func (this *MissingError) Error() string {
	return fmt.Sprintf("Error reading value for %q.", this.key)
}

type CastError struct {
	tipe  string
	value interface{}
}

func (this *CastError) Error() string {
	return fmt.Sprintf("Error casting value %q to %q.", this.value, this.tipe)
}

func IsCastError(err error) (ok bool) {
	_, ok = err.(*CastError)
	return
}

//GetString
func GetString(jobj map[string]interface{}, key string) (val string, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch ival.(type) {
	case string:
		val = ival.(string)
	default:
		val = fmt.Sprint(ival)
	}
	return
}

//GetInt
func GetInt(jobj map[string]interface{}, key string) (val int, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch ival.(type) {
	case int64:
		val = int(ival.(int64))
	case int:
		val = ival.(int)
	case float64:
		val = int(ival.(float64))
	case string:
		val, err = strconv.Atoi(ival.(string))
	default:
		err = &CastError{"int", ival}
	}
	return
}

//GetInt64
func GetInt64(jobj map[string]interface{}, key string) (val int64, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch ival.(type) {
	case int64:
		val = ival.(int64)
	case int:
		val = int64(ival.(int))
	case float64:
		val = int64(ival.(float64))
	case string:
		val, err = strconv.ParseInt(ival.(string), 10, 64)
	default:
		err = &CastError{"int64", ival}
	}
	return
}

//GetID
func GetId(jobj map[string]interface{}, key string) (id bson.ObjectId, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch ival.(type) {
	case bson.ObjectId:
		id = ival.(bson.ObjectId)
	case string:
		id, err = ReadId(ival.(string))
	default:
		err = &CastError{"id", ival}
	}
	return
}

//GetM
func GetM(jobj map[string]interface{}, key string) (val bson.M, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = &MissingError{key}
		return
	}
	switch ival.(type) {
	case bson.M:
		val = ival.(bson.M)
	default:
		err = &CastError{"bson.M", ival}
	}
	return
}

//GetBytes
func GetBytes(jobj map[string]interface{}, key string) ([]byte, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, &MissingError{key}
	}
	return toBytes(ival)
}

//GetStrings
func GetStrings(jobj map[string]interface{}, key string) ([]string, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, &MissingError{key}
	}
	return toStrings(ival)
}

//toBytes
func toBytes(ival interface{}) ([]byte, error) {
	val, ok := ival.([]byte)
	if !ok {
		return nil, &CastError{"[]byte", ival}
	}
	return val, nil
}

//toStrings
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

func ReadId(idStr string) (id bson.ObjectId, err error) {
	if !bson.IsObjectIdHex(idStr) {
		err = &CastError{"bson.ObjectId", idStr}
	} else{
		id = bson.ObjectIdHex(idStr)
	}
	return
}