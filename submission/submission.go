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
	Date     int64         "date"
	Results *bson.M "results"
}



func (f *File) IsSource() bool {
	return f.FileType == "SOURCE"
}

func NewFile(subId bson.ObjectId, fname, ftype string, data []byte) *File{
	now := time.Now().UnixNano()	
	fileId := bson.NewObjectId()
	return &File{fileId, subId, fname, ftype, data, now, new(bson.M)}
}
