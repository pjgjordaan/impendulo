package util

import (
	"fmt"
	"labix.org/v2/mgo/bson"
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
	val, ok := ival.(string)
	if !ok {
		return "", &CastError{"string", ival}
	}
	return val, nil
}

//GetInt
func GetInt(jobj map[string]interface{}, key string) (int, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, &MissingError{key}
	}
	val, ok := ival.(float64)
	if !ok {
		return -1, &CastError{"int", ival}
	}
	return int(val), nil
}

//GetInt64
func GetInt64(jobj map[string]interface{}, key string) (int64, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, &MissingError{key}
	}
	val, ok := ival.(float64)
	if !ok {
		return -1, &CastError{"int64", ival}
	}
	return int64(val), nil
}

//GetID
func GetID(jobj map[string]interface{}, key string) (bson.ObjectId, error) {
	ival, ok := jobj[key]
	if !ok {
		return bson.NewObjectId(), &MissingError{key}
	}
	val, ok := ival.(bson.ObjectId)
	if !ok {
		return bson.NewObjectId(), &CastError{"bson.ObjectId", ival}
	}
	return val, nil
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
