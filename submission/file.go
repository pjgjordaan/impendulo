package submission

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"strings"
)


//File stores a single file's data from a submission. 
type File struct {
	Id      bson.ObjectId "_id"
	SubId   bson.ObjectId "subid"
	Info    bson.M        "info"
	Data    []byte        "data"
	Results bson.M        "results"
}

//NewFile
func NewFile(subId bson.ObjectId, info map[string]interface{}, data []byte) *File {
	id := bson.NewObjectId()
	return &File{id, subId, info, data, bson.M{}}
}

//Type
func (f *File) Type() string {
	return f.InfoStr(TYPE)
}

//InfoStr retrieves file metadata.
func (f *File) InfoStr(key string) (val string) {
	val, _ = f.Info[key].(string)
	return val
}

func (this *File) Equals(that *File) bool {
	return reflect.DeepEqual(this, that)
}

//ParseName retrieves file metadata encoded in a file name.
//These file names must have the format:
//[[<package descriptor>"_"]*<file name>"_"]<time in nanoseconds>"_"<file number in current submission>"_"<modification char>
//Where values between '[]' are optional, '*' indicates 0 to many, values inside '""' are literals and values inside '<>' 
//describe the contents at that position.  
func ParseName(name string) (map[string]interface{}, error) {
	elems := strings.Split(name, "_")
	if len(elems) < 3 {
		return nil, fmt.Errorf("Encoded name %q does not have enough parameters.", name)
	}
	info := make(map[string]interface{})
	info[MOD] = elems[len(elems)-1]
	num, err := strconv.Atoi(elems[len(elems)-2])
	if err != nil {
		return nil, fmt.Errorf("%q in name %q could not be parsed as an int.", elems[len(elems)-2], name)
	}
	info[NUM] = num
	time, err := strconv.ParseInt(elems[len(elems)-3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%q in name %q could not be parsed as an int64.", elems[len(elems)-3], name)
	}
	info[TIME] = time
	fname := ""
	if len(elems) > 3 {
		info[NAME] = elems[len(elems)-4]
		fname = elems[len(elems)-4]
		pkg := ""
		for i := 0; i < len(elems)-4; i++ {
			pkg += elems[i]
			if i < len(elems)-5 {
				pkg += "."
			}
			if isOutFolder(elems[i]) {
				pkg = ""
			}
		}
		info[PKG] = pkg
	}
	if strings.HasSuffix(fname, JSRC) {
		info[TYPE] = SRC
	} else if strings.HasSuffix(fname, JCOMP) {
		info[TYPE] = EXEC
	} else {
		info[TYPE] = CHANGE
	}
	return info, nil
}
