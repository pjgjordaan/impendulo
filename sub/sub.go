package sub
import(
	"labix.org/v2/mgo/bson"
"time"
"strings"
"strconv"
)

/*
A struct used to store information about individual project submissions
in the database.
*/
type Submission struct {
	Id      bson.ObjectId "_id"
	Project string        "project"
	User    string        "user"
	Time    int64         "time"
	Mode    string        "mode"
}

func (s *Submission) IsTest() bool {
	return s.Mode == "TEST"
}

func NewSubmission(project, user, mode string) *Submission{
	subId := bson.NewObjectId()
	now := time.Now().UnixNano()	
	return &Submission{subId, project, user, now, mode}
}

func ReadSubmission(smap bson.M)*Submission{
	id := smap["_id"].(bson.ObjectId)
	proj := smap["project"].(string)
	usr := smap["user"].(string)
	time := smap["time"].(int64)
	mode := smap["mode"].(string)
	return &Submission{id, proj, usr, time, mode}
}

type File struct {
	Id       bson.ObjectId "_id"
	SubId bson.ObjectId "subid"
	Info bson.M "info"
	Data     []byte        "data"
	Results []interface{} "results"
}


func ReadFile(fmap bson.M)*File{
	id := fmap["_id"].(bson.ObjectId)
	subid := fmap["subid"].(bson.ObjectId)
	info := fmap["info"].(bson.M)
	data := fmap["data"].([]byte)
	res := fmap["results"].([]interface{})
	return &File{id, subid, info, data, res}
}

func NewFile(subId bson.ObjectId, info map[string] interface{}, data []byte)*File{
	id := bson.NewObjectId()
	return &File{id, subId, info, data, make([]interface{},0)}
}

func (f *File) Type() string{
	return f.InfoStr(TYPE)
}

func (f *File) InfoStr(key string)(val string){
	val,_ = f.Info[key].(string)
	return val
} 

func ParseName(name string)(info map[string] interface{}){
	info = make(map[string] interface{})
	elems := strings.Split(name, "_")
	info[MOD] = elems[len(elems) - 1]
	num,_ := strconv.Atoi(elems[len(elems) - 2])
	info[NUM] = num
	time,_ := strconv.ParseInt(elems[len(elems) - 3], 10, 64)
	info[TIME] = time
	if len(elems) > 3{
		info[NAME] = elems[len(elems) - 4]
		pkg := ""
		for i := 0; i < len(elems)-4; i ++{
			pkg += elems[i]
			if i < len(elems)-4{
				pkg += "."
			}
		}
		info[PKG] = pkg
	}
	return info
}


/*
func getFileInfo(info string)(f *File, err error){
	params := strings.Split(fname, "_")
	mod := char(params[len(params)-1])
	num,_ := strconv.Atoi(params[len(params)-2])
	modtime,_ := strconv.ParseInt(params[len(params)-3], 10, 64)
	fname := strings.Split(params[len(params)-4], ".")
	name := fname[0]
	ext := fname[1]
	pkg := params[len(params)-5]
}
*/

//File metadata
const(
	TYPE = "type"
SRC = "src"
EXEC = "exec"
CHANGE = "change"
TEST = "test"
ARCHIVE = "archive"
FTYPE = "ftype"
NAME = "name"
PKG = "pkg"
NUM = "num"
TIME = "time"
MOD = "mod"
)
