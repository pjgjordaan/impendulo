package submission
import(
	"labix.org/v2/mgo/bson"
"time"
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

/*
A struct used to store individual files in the database.
*/
type File struct {
	Id       bson.ObjectId "_id"
	SubId    bson.ObjectId "subid"
	Name     string        "name"
	FileType string        "type"
	Data     []byte        "data"
	Time     int64         "time"
	Number int "number"
	Modification char "modification"
	Results *bson.M "results"
}



func (f *File) IsSource() bool {
	return f.FileType == "SOURCE"
}

func NewFile(subId bson.ObjectId, fname, ftype string, data []byte) *File{
	//Specific to how the file names are formatted currently, should change.	
	params := strings.Split(fname, "_")
	mod := char(params[len(params)-1])
	num,_ := strconv.Atoi(params[len(params)-2])
	modtime,_ := strconv.ParseInt(params[len(params)-3], 10, 64)
	fname := strings.Split(params[len(params)-4], ".")
	name := fname[0]
	ext := fname[1]
	pkg := params[len(params)-5]
	fileId := bson.NewObjectId()
	return &File{fileId, subId, name, ftype, data, modtime, num, mod, new(bson.M)}
}
