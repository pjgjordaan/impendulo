package project

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"strings"
)

//File stores a single file's data from a submission.
type File struct {
	Id       bson.ObjectId "_id"
	SubId    bson.ObjectId "subid"
	Name     string        "name"
	Package  string        "package"
	Type     string        "type"
	FileType string        "ftype"
	Mod      string        "mod"
	Num      int           "num"
	Time     int64         "time"
	Data     []byte        "data"
	Results  bson.M        "results"
}

func (this *File) TypeName() string {
	return "file"
}

func (this *File) String() string {
	return "Type: project.File; Id: " + this.Id.Hex() + "; SubId: " + this.SubId.Hex() + "; Name: " + this.Name + "; Package: " + this.Package + "; Type: " + this.Type + "; FileType: " + this.FileType + "; Mod: " + this.Mod + "; Num: " + strconv.Itoa(this.Num) + "; Time: " + util.Date(this.Time)
}

func (this *File) SetMod(mod string) {
	switch mod {
	case "b":
		this.Mod = "Compiled"
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
	if reflect.DeepEqual(this, that) {
		return true
	}
	return that != nil && this.String() == that.String() && bytes.Equal(this.Data, that.Data)
}

func (this *File) Same(that *File) bool {
	return this.Id == that.Id
}


func (this *File) CanProcess() bool {
	return this.Type == SRC || this.Type == ARCHIVE
}

//NewFile
func NewFile(subId bson.ObjectId, info map[string]interface{}, data []byte) (file *File, err error) {
	id := bson.NewObjectId()
	file = &File{Id: id, SubId: subId, Data: data}
	//Non essential fields
	file.Type, err = util.GetString(info, TYPE)
	if err != nil && util.IsCastError(err) {
		return
	}
	file.FileType, err = util.GetString(info, FTYPE)
	if err != nil && util.IsCastError(err) {
		return
	}
	//Essential fields
	file.Name, err = util.GetString(info, NAME)
	if err != nil {
		return
	}
	file.Package, err = util.GetString(info, PKG)
	if err != nil {
		return
	}
	mod, err := util.GetString(info, MOD)
	if err != nil {
		return
	}
	file.SetMod(mod)
	file.Num, err = util.GetInt(info, NUM)
	if err != nil {
		return
	}
	file.Time, err = util.GetInt64(info, TIME)
	return
}

//NewFile
func NewArchive(subId bson.ObjectId, data []byte, ftype string) *File {
	id := bson.NewObjectId()
	return &File{Id: id, SubId: subId, Data: data, FileType: ftype, Type: ARCHIVE}
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
