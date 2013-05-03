package sub

import (
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
	"time"
)

/*
Individual project submissions
*/
type Submission struct {
	Id      bson.ObjectId "_id"
	Project string        "project"
	User    string        "user"
	Time    int64         "time"
	Mode    string        "mode"
	Lang string "lang"
}


func (s *Submission) IsTest() bool {
	return s.Mode == TEST_MODE
}

func NewSubmission(project, user, mode, lang string) *Submission {
	subId := bson.NewObjectId()
	now := time.Now().UnixNano()
	return &Submission{subId, project, user, now, mode, lang}
}

/*
Extract submission from a mongo map.
*/
func ReadSubmission(smap bson.M) (*Submission, error) {
	id, err := utils.GetID(smap, ID) 
	if err != nil{
		return nil, err
	}
	proj, err := utils.GetString(smap, PROJECT)
	if err != nil{
		return nil, err
	}
	usr, err := utils.GetString(smap, USER)
	if err != nil{
		return nil, err
	}
	time, err := utils.GetInt64(smap, TIME)
	if err != nil{
		return nil, err
	}
	mode, err := utils.GetString(smap, MODE)
	if err != nil{
		return nil, err
	}
	lang, err := utils.GetString(smap, LANG)
	if err != nil{
		return nil, err
	}
	return &Submission{id, proj, usr, time, mode, lang}, nil
}

/*
Single file's data from a submission. 
*/
type File struct {
	Id      bson.ObjectId "_id"
	SubId   bson.ObjectId "subid"
	Info    bson.M        "info"
	Data    []byte        "data"
	Results bson.M "results"
}

/*
Extract file data from a mongo map
*/
func ReadFile(fmap bson.M)(*File, error) {
	id, err := utils.GetID(fmap, ID)
	if err != nil{
		return nil, err
	}
	subid, err := utils.GetID(fmap, SUBID)
	if err != nil{
		return nil, err
	}
	info, err := utils.GetM(fmap, INFO)
		if err != nil{
		return nil, err
	}
	data, err := utils.GetBytes(fmap, DATA)
		if err != nil{
		return nil, err
	}
	res, err := utils.GetM(fmap, RES)
	if err != nil{
		return nil, err
	}
	return &File{id, subid, info, data, res}, nil
}

func NewFile(subId bson.ObjectId, info map[string]interface{}, data []byte) *File {
	id := bson.NewObjectId()
	return &File{id, subId, info, data, bson.M{}}
}

func (f *File) Type() string {
	return f.InfoStr(TYPE)
}

/*
Retrieve file metadata from a mongo map
*/
func (f *File) InfoStr(key string) (val string) {
	val, _ = f.Info[key].(string)
	return val
}

/*
 Retrieve file metadata encoded in a file name. These file names must have the format:
 [[<package descriptor>"_"]*<file name>"_"]<time in nanoseconds>"_"<file number in current submission>"_"<modification char>
 Where values between '[]' are optional, '*' indicates 0 to many, values inside '""' are literals and values inside '<>' 
 describe the contents at that position.  
*/
func ParseName(name string) (info map[string]interface{}) {
	info = make(map[string]interface{})
	elems := strings.Split(name, "_")
	info[MOD] = elems[len(elems)-1]
	num, _ := strconv.Atoi(elems[len(elems)-2])
	info[NUM] = num
	time, _ := strconv.ParseInt(elems[len(elems)-3], 10, 64)
	info[TIME] = time
	if len(elems) > 3 {
		info[NAME] = elems[len(elems)-4]
		pkg := ""
		for i := 0; i < len(elems)-4; i++ {
			pkg += elems[i]
			if i < len(elems)-4 {
				pkg += "."
			}
		}
		info[PKG] = pkg
	}
	return info
}

/*
Constants used client and server-side to describe file and
submission data
*/
const (
	ID = "_id"
	PROJECT = "project"
	USER = "user"
	TIME = "time"
	MODE = "mode"
	TYPE    = "type"
	SRC     = "src"
	EXEC    = "exec"
	CHANGE  = "change"
	ARCHIVE = "archive"
	TEST = "test"
	FILE_MODE = "file_remote"
	TEST_MODE    = "archive_test"
	ARCHIVE_MODE = "archive_remote"
	FTYPE   = "ftype"
	NAME    = "name"
	PKG     = "pkg"
	NUM     = "num"
	MOD     = "mod"
	LANG = "lang"
	SUBID = "subid"
	INFO = "info"
	DATA = "data"
	RES = "results"
)