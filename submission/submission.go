package submission
import(
	"labix.org/v2/mgo/bson"
"time"
)

const(
	SOURCE = iota
CHANGE
EXECUTABLE
ARCHIVE
TEST
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

type File interface {
	Type() int
}

type Archive struct{
	Info *FileInfo "info"
	Data     []byte        "data"
}

func (a *Archive) int{
	return ARCHIVE
}			


type Test struct {
	Id       bson.ObjectId "_id"
	Info *FileInfo "info"
	Data     []byte        "data"
}

func (t *Test)  Type() int{
	return Test
}

type Source struct {
	Id       bson.ObjectId "_id"
	Info *FileInfo "info"
	Data     []byte        "data"
	Results *bson.M "results"
}

func (s *Source)  Type() int{
	return SOURCE
}

type Exec struct {
	Id       bson.ObjectId "_id"
	Info *FileInfo "info"
	Data     []byte        "data"
}

func (e *Exec)  Type() int{
	return EXECUTABLE
}


type Change struct {
	Id       bson.ObjectId "_id"
	Info *FileInfo "info"
}

func (c *Change)  Type() int{
	return CHANGE
}


func NewSource(finfo *FileInfo) *Source{
	fileId := bson.NewObjectId()
	return &Source{fileId, finfo, nil, new(bson.M)}
}

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


type FileInfo struct{
	SubId    bson.ObjectId "subid"
	Name     string        "name"
	Package string "package"
	FileType string        "type"
	Modification string "modification"
	Number string "number"
	Time     string         "time"
}

func NewTestInfo(subId bson.ObjectId, name, ftype string, time int64)(fi *FileInfo){
	return &FileInfo{SubId: subId,Name: name, FileType: ftype, Time: time}
}

func NewTest(subId bson.ObjectId, name, ftype string, time int64)*Test{
	return &Test{FileInfo: newTestInfo(subId,name, ftype, time)}
}


func NewArchiveInfo(ftype string)(fi *FileInfo){
	return &FileInfo{FileType: ftype}
}

func NewArchive(ftype string) *Archive{
	return &Archive{FileInfo: newArchiveInfo(ftype)}
}

func NewFileInfo(subId bson.ObjectId, name, pkg, ftype, mod string, num int, time int64)(fi *FileInfo){
	return &FileInfo{subId, name,pkg, ftype, mod, num, time}
}

func NewSource(subId bson.ObjectId, name, pkg, ftype, mod string, num int, time int64)*Source{
	return &Source{FileInfo: NewFileInfo(subId,name,pkg,ftype,mod,num, time)}
}


func NewExec(subId bson.ObjectId, name, pkg, ftype, mod string, num int, time int64)*Exec{
	return &Exec{FileInfo: NewFileInfo(subId,name,pkg,ftype,mod,num, time)}
}


func NewChange(subId bson.ObjectId, name, pkg, ftype, mod string, num int, time int64)*Change{
	return &Change{FileInfo: NewFileInfo(subId,name,pkg,ftype,mod,num, time)}
}

