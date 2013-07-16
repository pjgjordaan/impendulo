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
func GetString(jobj map[string]interface{}, key string) (string, error) {
	ival, ok := jobj[key]
	if !ok {
		return "", &MissingError{key}
	}
	switch ival.(type) {
	case string:
		return readString(ival), nil
	default:
		return fmt.Sprint(ival), nil
	}
}

func readString(ival interface{}) (string) {
	val, _ := ival.(string)
	return val
}

//GetInt
func GetInt(jobj map[string]interface{}, key string) (int, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, &MissingError{key}
	}
	switch ival.(type) {
	case int64:
		return int(readInt64(ival)), nil
	case int:
		return readInt(ival), nil
	case float64:
		return int(readFloat64(ival)), nil
	case string:
		return strconv.Atoi(readString(ival))	
	default:
		return -1, &CastError{"int", ival}
	}	
	
}

func readInt(ival interface{}) (int) {
	val, _ := ival.(int)
	return val
}

func readFloat64(ival interface{}) (float64) {
	val, _ := ival.(float64)
	return val
}

//GetInt64
func GetInt64(jobj map[string]interface{}, key string) (int64, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, &MissingError{key}
	}
	switch ival.(type) {
	case int64:
		return readInt64(ival), nil
	case int:
		return int64(readInt(ival)), nil
	case float64:
		return int64(readFloat64(ival)), nil
	case string:
		return strconv.ParseInt(readString(ival), 10, 64)
	default:
		return -1, &CastError{"int64", ival}
	}	
}

func readInt64(ival interface{})(int64){
	val, _ := ival.(int64)
	return val
}

//GetID
func GetID(jobj map[string]interface{}, key string) (bson.ObjectId, error) {
	ival, ok := jobj[key]
	if !ok {
		return bson.NewObjectId(), &MissingError{key}
	}
	switch ival.(type) {
	case bson.ObjectId:
		return ival.(bson.ObjectId), nil
	case string:
		idStr := readString(ival)
		if bson.IsObjectIdHex(idStr){
			return bson.ObjectIdHex(idStr), nil
		}
	}
	return bson.NewObjectId(), &CastError{"bson.ObjectId", ival}
}

//GetM
func GetM(jobj map[string]interface{}, key string) (bson.M, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, &MissingError{key}
	}
	val, ok := ival.(bson.M)
	if !ok {
		return nil, &CastError{"bson.M", ival}
	}
	return val, nil
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
