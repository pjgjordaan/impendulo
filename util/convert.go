package util

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"strconv"
)

func ReadId(idStr string) (id bson.ObjectId, err error) {
	if !bson.IsObjectIdHex(idStr) {
		err = &CastError{"bson.ObjectId", idStr}
	} else{
		id = bson.ObjectIdHex(idStr)
	}
	return
}

//GetString
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

//GetInt
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

//GetInt64
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
		err = &CastError{"int64", v}
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

//GetM
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
