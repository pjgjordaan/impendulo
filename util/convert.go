package util

import (
	"fmt"
	"labix.org/v2/mgo/bson"
)

//GetString
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

//GetInt
func GetInt(jobj map[string]interface{}, key string) (int, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(float64)
	if !ok {
		return -1, fmt.Errorf("%q could not be parsed as an int.", ival)
	}
	return int(val), nil
}

//GetInt64
func GetInt64(jobj map[string]interface{}, key string) (int64, error) {
	ival, ok := jobj[key]
	if !ok {
		return -1, fmt.Errorf("Error reading value for %q ", key)
	}
	val, ok := ival.(float64)
	if !ok {
		return -1, fmt.Errorf("%q could not be parsed as an int64.", ival)
	}
	return int64(val), nil
}

//GetID
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

//GetM
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

//GetBytes
func GetBytes(jobj map[string]interface{}, key string) ([]byte, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, fmt.Errorf("Error reading value for %q ", key)
	}
	return toBytes(ival)
}

//GetStrings
func GetStrings(jobj map[string]interface{}, key string) ([]string, error) {
	ival, ok := jobj[key]
	if !ok {
		return nil, fmt.Errorf("Error reading value for %q ", key)
	}
	return toStrings(ival)
}

//toBytes
func toBytes(ival interface{}) ([]byte, error) {
	val, ok := ival.([]byte)
	if !ok {
		return nil, fmt.Errorf("Error casting value %q to []byte", ival)
	}
	return val, nil
}

//toStrings
func toStrings(ivals interface{}) ([]string, error) {
	vals, ok := ivals.([]string)
	if !ok {
		islice, ok := ivals.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Error casting value %q to []string", ivals)
		}
		vals = make([]string, len(islice))
		for i, ival := range islice {
			val, ok := ival.(string)
			if !ok {
				return nil, fmt.Errorf("Error casting value %q to string", ival)
			}
			vals[i] = val
		}
	}
	return vals, nil
}
