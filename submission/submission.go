package submission

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//Submission is used for individual project submissions
type Submission struct {
	Id      bson.ObjectId "_id"
	Project string        "project"
	User    string        "user"
	Time    int64         "time"
	Mode    string        "mode"
	Lang    string        "lang"
}

//IsTest
func (s *Submission) IsTest() bool {
	return s.Mode == TEST_MODE
}

func (this *Submission) String() string {
	return "Project: " + this.Project + "; User: " + this.User + "; Time: " + time.Unix(0, this.Time).String()

}

func (this *Submission) Equals(that *Submission) bool {
	return reflect.DeepEqual(this, that)
}

//NewSubmission
func NewSubmission(project, user, mode, lang string) *Submission {
	subId := bson.NewObjectId()
	now := time.Now().UnixNano()
	return &Submission{subId, project, user, now, mode, lang}
}

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

//isOutFolder
func isOutFolder(arg string) bool {
	return arg == SRC_DIR || arg == BIN_DIR
}

//Constants used client and server-side to for submission data.
const (
	ID           = "_id"
	PROJECT      = "project"
	USER         = "user"
	TIME         = "time"
	MODE         = "mode"
	TYPE         = "type"
	SRC          = "src"
	EXEC         = "exec"
	CHANGE       = "change"
	ARCHIVE      = "archive"
	TEST         = "test"
	FILE_MODE    = "file_remote"
	TEST_MODE    = "archive_test"
	ARCHIVE_MODE = "archive_remote"
	FTYPE        = "ftype"
	NAME         = "name"
	PKG          = "pkg"
	NUM          = "num"
	MOD          = "mod"
	LANG         = "lang"
	SUBID        = "subid"
	INFO         = "info"
	DATA         = "data"
	RES          = "results"
	JSRC         = ".java"
	JCOMP        = ".class"
	BIN_DIR      = "bin"
	SRC_DIR      = "src"
)
