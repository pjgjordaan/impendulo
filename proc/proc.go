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

func (t *TestBuilder) Setup(project string) {
	t.m.Lock()
	if !t.Tests[project] {
		tests := GetTests(project)
		err := utils.Unzip(filepath.Join(t.TestDir, project), tests.Data)
		if err != nil {
			panic(err)
		}
		t.Tests[project] = true
	}
	t.m.Unlock()
}

func GetTests(project string) (tests *sub.File) {
	smap, err := db.GetMatching(db.SUBMISSIONS, bson.M{"project": project, "mode": "TEST"})
	if err != nil {
		panic(err)
	}
	s := sub.ReadSubmission(smap)
	fmap, err := db.GetMatching(db.FILES, bson.M{"subid": s.Id})
	if err != nil {
		panic(err)
	}
	tests = sub.ReadFile(fmap)
	return tests
}

var testBuilder *TestBuilder

func Serve(files chan *sub.File) {
	tools.AddTool(tools.FB)
	testBuilder = NewTestBuilder()
	// Start handlers
	for f := range files {
		go Process(f)
	}
}

func Process(f *sub.File) {
	t := f.Type()
	if t == sub.ARCHIVE {
		ProcessArchive(f)
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
}

func ProcessSource(src *sub.File) {
	ti := setupSource(src)
	if tools.Compile(src.Id, ti) {
		runTests(src, ti)
		tools.RunTools(src.Id, ti)
	}
	//clean up
	err := os.RemoveAll(ti.Dir)
	if err != nil {
		panic(err)
	}
}

func setupSource(src *sub.File) (ti *tools.TargetInfo) {
	dir := filepath.Join(os.TempDir(), src.Id.Hex())
	ti = tools.NewTarget(src.InfoStr(sub.NAME), src.InfoStr(sub.PKG), dir)
	err := utils.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), src.Data)
	if err != nil {
		panic(err)
	}
	return ti
}

func runTests(src *sub.File, ti *tools.TargetInfo) {
	smap, err := db.GetById(db.SUBMISSIONS, src.SubId)
	if err != nil {
		utils.Log("Tests not found:", err)
		return
	}
	s := sub.ReadSubmission(smap)
	testBuilder.Setup(s.Project)
	tools.RunTests(src.Id, s.Project, ti)
}

func ProcessExec(s *sub.File) {
}

func ProcessChange(s *sub.File) {
}

func ProcessTest(s *sub.File) {
}

func ProcessArchive(s *sub.File) {
	files, err := utils.ReadZip(s.Data)
	if err != nil {
		utils.Log("Bad archive: ", err)
		return
	}
	for name, data := range files {
		info := sub.ParseName(name)
		f := sub.NewFile(s.SubId, info, data)
		Process(f)
	}
}
