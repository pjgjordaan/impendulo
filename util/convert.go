package util

import (
	"fmt"
	"labix.org/v2/mgo/bson"
)

func GetString(jobj map[string]interface{}, key string) (string, error) {
	ival, ok := jobj[key]
	if !ok {
		return "", fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(string)
	if !ok {
		return "", fmt.Errorf("Error casting value %q to string", ival)
	}
	return val, nil
}

func GetInt(jobj map[string]interface{}, key string) (int, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(int)
	if !ok {
		return -1, fmt.Errorf("Error casting value %q to int", ival)
	}
	return val, nil
}

func GetInt64(jobj map[string]interface{}, key string) (int64, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(int64)
	if !ok {
		return -1, fmt.Errorf("Error casting value %q to int64", ival)
	}
	return val, nil
}

func GetID(jobj map[string]interface{}, key string) (bson.ObjectId, error) {
	ival, ok := jobj[key]
	if !ok {
		return bson.NewObjectId(), fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(bson.ObjectId)
	if !ok {
		return bson.NewObjectId(), fmt.Errorf("Error casting value %q to bson.ObjectId", ival)
	}
	return val, nil
}

func GetM(jobj map[string]interface{}, key string) (bson.M, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(bson.M)
	if !ok {
		return nil, fmt.Errorf("Error casting value %q to bson.M", ival)
	}
	return val, nil
}

func GetBytes(jobj map[string]interface{}, key string) ([]byte, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, fmt.Errorf("Error reading value for %q ", key)
	}
	return toBytes(ival)
}

func GetStrings(jobj map[string]interface{}, key string) ([]string, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, fmt.Errorf("Error reading value for %q ", key)
	}
	return toStrings(ival)
}

func toBytes(ival interface{}) ([]byte, error) {
	val, ok := ival.([]byte)
	if !ok {
		return nil, fmt.Errorf("Error casting value %q to []byte", ival)
	}
	return val, nil
}

func toStrings(ivals interface{}) ([]string, error) {
	vals, ok := ivals.([]string)
	if !ok {
		return nil, fmt.Errorf("Error casting value %q to []string", ivals)
	}
	return vals, nil
}

func MEqual(m1, m2 bson.M) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}
	return true
}

func StringsEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for k, v := range s1 {
		if s2[k] != v {
			return false
		}
	}
	return true
}
