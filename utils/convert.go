package utils

import (
	"errors"
	"labix.org/v2/mgo/bson"
)

func GetString(jobj map[string]interface{}, key string) (val string, err error) {
	ival, ok := jobj[key]
	
	if ok {
		val, ok = ival.(string)
	}
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return val, err
}

func GetInt(jobj map[string]interface{}, key string) (val int, err error) {
	ival, ok := jobj[key]
	if ok {
		val, ok = ival.(int)
	}
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return val, err
}

func GetInt64(jobj map[string]interface{}, key string) (val int64, err error) {
	ival, ok := jobj[key]
	if ok {
		val, ok = ival.(int64)
	}
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return val, err
}

func GetID(jobj map[string]interface{}, key string) (val bson.ObjectId, err error) {
	ival, ok := jobj[key]
	if ok {
		val, ok = ival.(bson.ObjectId)
	}
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return val, err
}

func GetM(jobj map[string]interface{}, key string) (val bson.M, err error) {
	ival, ok := jobj[key]
	if ok {
		val, ok = ival.(bson.M)
	}
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return val, err
}

func GetBytes(jobj map[string]interface{}, key string) (val []byte, err error) {
	ival, ok := jobj[key]
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return ToBytes(ival)
}

func GetStrings(jobj map[string]interface{}, key string) ([]string, error) {
	ival, ok := jobj[key]
	if !ok{
		return nil, errors.New("Error reading value for: " + key)
	}
	return ToStrings(ival)
}

func ToBytes(bint interface{})([]byte, error){
	val, ok := bint.([]byte)
	if !ok{
		return nil, errors.New("Error casting interface to []byte")
	}
	return val, nil
}


func ToStrings(sint interface{})([]string, error){
	ivals, ok := sint.([]string)
	if !ok{
		return nil, errors.New("Error casting interface to []interface")
	}
/*	vals := make([]string, len(ivals))
	for i, ival := range ivals{
		val, ok := ival.(string)
		if !ok{
			return nil, errors.New("Error casting interface to string")
		}
		vals[i] = ival
	}*/
	return ivals, nil
}

func MEqual(m1, m2 bson.M) bool{
	if len(m1) != len(m2){
		return false
	}
	for k, v := range m1{
		if m2[k] != v{
			return false
		}
	}
	return true
}

func StringsEqual(s1, s2 []string) bool{
	if len(s1) != len(s2){
		return false
	}
	for k, v := range s1{
		if s2[k] != v{
			return false
		}
	}
	return true
}
