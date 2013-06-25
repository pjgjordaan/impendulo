package project

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
	Name string "name"
	Package string "package"
	Type string "type"
	FileType string "type"
	Mod string "mod"
	Num int "num"
	Time int64 "time"
	Data    []byte        "data"
	Results bson.M        "results"
}

//NewFile
func NewFile(subId bson.ObjectId, info map[string]interface{}, data []byte) (*File, error) {
	id := bson.NewObjectId()
	file := &File{Id: id, SubId: subId, Data: data}
	if v, ok := info[NAME]; ok{
		file.Name, ok = v.(string)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as a string.", v)
		}
	}
	if v, ok := info[PKG]; ok{
		file.Package, ok = v.(string)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as a string.", v)
		}
	}
	if v, ok := info[TYPE]; ok{
		file.Type, ok = v.(string)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as a string.", v)
		}
	}
	if v, ok := info[FTYPE]; ok{
		file.FileType, ok = v.(string)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as a string.", v)
		}
	}
	if v, ok := info[MOD]; ok{
		mod, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as a string.", v)
		}
		file.SetMod(mod)
	}
	if v, ok := info[NUM]; ok{
		file.Num, ok = v.(int)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as an int.", v)
		}
	}
	if v, ok := info[TIME]; ok{
		file.Time, ok = v.(int64)
		if !ok {
			return nil, fmt.Errorf("%q could not be parsed as an int64.", v)
		}
	}
	return file, nil
}

//NewFile
func NewArchive(subId bson.ObjectId, data []byte, ftype string) *File{
	id := bson.NewObjectId()
	return &File{Id: id, SubId: subId, Data: data, FileType: ftype, Type: ARCHIVE}
}


func (this *File) SetMod(mod string) {
	switch mod {
	case "c":
		this.Mod = "Saved"
	case "r":
		this.Mod = "Removed"
	case "l":
		this.Mod = "Launched"
	case "f":
		this.Mod = "From"
	case "t":
		this.Mod = "To"
	case "a":
		this.Mod = "Added"
	default:
		this.Mod = "Unknown"
	}
}

func (this *File) Equals(that *File) bool {
	return reflect.DeepEqual(this, that)
}

//ParseName retrieves file metadata encoded in a file name.
//These file names must have the format:
//[[<package descriptor>"_"]*<file name>"_"]<time in nanoseconds>"_"<file number in current submission>"_"<modification char>
//Where values between '[]' are optional, '*' indicates 0 to many, values inside '""' are literals and values inside '<>'
//describe the contents at that position.
func ParseName(name string) (*File, error) {
	elems := strings.Split(name, "_")
	if len(elems) < 3 {
		return nil, fmt.Errorf("Encoded name %q does not have enough parameters.", name)
	}
	file := new(File)
	file.Id = bson.NewObjectId()
	var err error
	file.SetMod(elems[len(elems)-1])
	file.Num, err = strconv.Atoi(elems[len(elems)-2])
	if err != nil {
		return nil, fmt.Errorf("%q in name %q could not be parsed as an int.", elems[len(elems)-2], name)
	}
	file.Time, err = strconv.ParseInt(elems[len(elems)-3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%q in name %q could not be parsed as an int64.", elems[len(elems)-3], name)
	}
	if len(elems) > 3 {
		file.Name = elems[len(elems)-4]
		for i := 0; i < len(elems)-4; i++ {
			file.Package += elems[i]
			if i < len(elems)-5 {
				file.Package += "."
			}
			if isOutFolder(elems[i]) {
				file.Package = ""
			}
		}
	}
	if strings.HasSuffix(file.Name, JSRC) {
		file.Type = SRC
		file.FileType = JAVA
	} else if strings.HasSuffix(file.Name, JCOMP) {
		file.Type = EXEC
		file.FileType = CLASS
	} else {
		file.Type = CHANGE
		file.FileType = EMPTY
	}
	return file, nil
}
