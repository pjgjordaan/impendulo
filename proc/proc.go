package proc

import (
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/sub"
	"github.com/disco-volante/intlola/tools"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"sync"
"fmt"
)

type TestBuilder struct {
	Tests   map[string]bool
	m       *sync.Mutex
	TestDir string
}

func NewTestBuilder() *TestBuilder {
	dir := filepath.Join(os.TempDir(), "tests")
	return &TestBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

func (t *TestBuilder) Setup(project string)(ret bool){
	t.m.Lock()
	if !t.Tests[project] {
		tests, err := GetTests(project)
		if err == nil{
			err := utils.Unzip(filepath.Join(t.TestDir, project), tests.Data)
			if err != nil {
				panic(err)
			}
			t.Tests[project] = true
			ret = true
		} 
	}
	t.m.Unlock()
	return ret
}

func GetTests(project string) (tests *sub.File, err error) {
	smap, err := db.GetMatching(db.SUBMISSIONS, bson.M{sub.PROJECT: project, sub.MODE: sub.TEST})
	if err != nil {
		return tests, err
	}
	s := sub.ReadSubmission(smap)
	fmap, err := db.GetMatching(db.FILES, bson.M{"subid": s.Id})
	if err != nil {
		return tests, err
	}
	tests = sub.ReadFile(fmap)
	return tests, err
}

var testBuilder *TestBuilder

func Serve(files chan *sub.File) {
	tools.AddTool(tools.FB)
	testBuilder = NewTestBuilder()
	// Start handlers
	status := make(chan string)
	go StatusListener(status)
	for f := range files {
		fmt.Println(f)
		go Process(f, status)
	}
}

func StatusListener(status chan string){
	for stat := range status{
		fmt.Println(stat)
	}
}

func Process(f *sub.File, status chan string) {
	status <- "Processing: "+f.InfoStr(sub.NAME)
	t := f.Type()
	if t == sub.ARCHIVE {
		ProcessArchive(f, status)
	} else {
		err := db.AddSingle(db.FILES, f)
		if err != nil {
			panic(err)
		}
		if t == sub.SRC {
			ProcessSource(f)
		} else if t == sub.EXEC {
			ProcessExec(f)
		} else if t == sub.CHANGE {
			ProcessChange(f)
		} else if t == sub.TEST {
			ProcessTest(f)
		}
	}
	status <- "Processed: "+f.InfoStr(sub.NAME)
}

func ProcessSource(src *sub.File) {
	smap, err := db.GetById(db.SUBMISSIONS, src.SubId)
	if err != nil {
		utils.Log("Submission not found:", err)
		return
	}
	s := sub.ReadSubmission(smap)
	ti := setupSource(src, s)
	if tools.Compile(src.Id, ti) {
		runTests(src, s.Project, ti)
		tools.RunTools(src.Id, ti)
	}
	//clean up
	err = os.RemoveAll(ti.Dir)
	if err != nil {
		panic(err)
	}
}

func setupSource(src *sub.File, s *sub.Submission) (ti *tools.TargetInfo) {
	dir := filepath.Join(os.TempDir(), src.Id.Hex())
	ti = tools.NewTarget(src.InfoStr(sub.NAME),s.Lang, src.InfoStr(sub.PKG), dir)
	err := utils.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), src.Data)
	if err != nil {
		panic(err)
	}
	return ti
}

func runTests(src *sub.File, project string, ti *tools.TargetInfo) {
	if testBuilder.Setup(project){
		tools.RunTests(src.Id, project, ti)
	} else {
		utils.Log("No tests found for ", project)
	}
}

func ProcessExec(s *sub.File) {
}

func ProcessChange(s *sub.File) {
}

func ProcessTest(s *sub.File) {
}

func ProcessArchive(s *sub.File, status chan string) {
	files, err := utils.ReadZip(s.Data)
	if err != nil {
		utils.Log("Bad archive: ", err)
		return
	}
	for name, data := range files {
		info := sub.ParseName(name)
		f := sub.NewFile(s.SubId, info, data)
		Process(f, status)
	}
}
