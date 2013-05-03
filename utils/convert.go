package utils
import(
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


func GetBytes(jobj map[string]interface{}, key string) (val [] byte, err error) {
	ival, ok := jobj[key]
	if ok {
		val, ok = ival.([] byte)
	}
	if !ok {
		err = errors.New("Error reading value for: " + key)
	}
	return val, err
}