package db

import(
	"testing"
"github.com/godfried/cabanga/submission"
"github.com/godfried/cabanga/tool"
"labix.org/v2/mgo/bson"
)

func TestSetup(t *testing.T){
	Setup(DEFAULT_CONN)
	getSession().Close()
	Setup(TEST_CONN)
	s := getSession()
	defer s.Close()
	err := s.DB(TEST_DB).DropDatabase()
	if err != nil{
		t.Error(err)
	}
}

func TestRemoveFile(t *testing.T){
	Setup(TEST_CONN)
	s := getSession()
	defer s.Close()
	f :=  submission.NewFile(bson.NewObjectId(), map[string]interface{}{"a":"b"}, []byte("aa"))
	err := AddFile(f)
	if err != nil{
		t.Error(err)
	}
	err = RemoveFileByID(f.Id)
	if err != nil{
		t.Error(err)
	}
	matcher := bson.M{"_id":f.Id}
	f, err = GetFile(matcher)
	if f != nil || err == nil{
		t.Error("File not deleted")
	}
	err = s.DB(TEST_DB).DropDatabase()
	if err != nil{
		t.Error(err)
	}
}

func TestGetFile(t *testing.T){
	Setup(TEST_CONN)
	s := getSession()
	defer s.Close()
	f :=  submission.NewFile(bson.NewObjectId(), map[string]interface{}{"a":"b"}, []byte("aa"))
	err := AddFile(f)
	if err != nil{
		t.Error(err)
	}
	matcher := bson.M{"_id":f.Id}
	dbFile, err := GetFile(matcher)
	if err != nil{
		t.Error(err)
	}
	if !f.Equals(dbFile){
		t.Error("Files not equivalent")
	}
	err = s.DB(TEST_DB).DropDatabase()
	if err != nil{
		t.Error(err)
	}
}

func TestGetSubmission(t *testing.T){
	Setup(TEST_CONN)
	s := getSession()
	defer s.Close()
	sub :=  submission.NewSubmission("project", "user", submission.FILE_MODE, "java")
	err := AddSubmission(sub)
	if err != nil{
		t.Error(err)
	}
	matcher := bson.M{"_id":sub.Id}
	dbSub, err := GetSubmission(matcher)
	if err != nil{
		t.Error(err)
	}
	if !sub.Equals(dbSub){
		t.Error("Submissions not equivalent")
	}
	err = s.DB(TEST_DB).DropDatabase()
	if err != nil{
		t.Error(err)
	}
}

func TestGetTool(t *testing.T){
	Setup(TEST_CONN)
	s := getSession()
	defer s.Close()
	fb :=  &tool.Tool{bson.NewObjectId(), "findbugs", tool.JAVA, "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", tool.WARNS, []string{tool.JAVA, "-jar"}, []string{"-textui", "-low"}, bson.M{}, tool.PKG_PATH}
	err := AddTool(fb)
	if err != nil{
		t.Error(err)
	}
	matcher := bson.M{"_id":fb.Id}
	dbTool, err := GetTool(matcher)
	if err != nil{
		t.Error(err)
	}
	if !fb.Equals(dbTool){
		t.Error("Tools not equivalent")
	}
	err = s.DB(TEST_DB).DropDatabase()
	if err != nil{
		t.Error(err)
	}
}


func TestGetTools(t *testing.T){
	Setup(TEST_CONN)
	s := getSession()
	defer s.Close()
	fb :=  &tool.Tool{bson.NewObjectId(), "findbugs", tool.JAVA, "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", tool.WARNS, []string{tool.JAVA, "-jar"}, []string{"-textui", "-low"}, bson.M{}, tool.PKG_PATH}
	javac := &tool.Tool{bson.NewObjectId(), tool.COMPILE, tool.JAVA, tool.JAVAC, tool.WARNS, tool.ERRS, []string{}, []string{"-implicit:class"}, bson.M{tool.CP: ""}, tool.FILE_PATH}
	tools := []*tool.Tool{fb, javac}
	err := AddTool(fb)
	if err != nil{
		t.Error(err)
	}
	err = AddTool(javac)
	if err != nil{
		t.Error(err)
	}
	matcher := bson.M{tool.LANG:tool.JAVA}
	dbTools, err := GetTools(matcher)
	if err != nil{
		t.Error(err)
	}
	for _, t0 := range tools{ 
		found := false
		for _, t1 := range dbTools{
			if t0.Equals(t1){
				found = true
				break
			}
		}
		if !found{
			t.Error("No match found", t0)
		}
	}
	err = s.DB(TEST_DB).DropDatabase()
	if err != nil{
		t.Error(err)
	}
}