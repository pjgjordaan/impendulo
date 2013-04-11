package proc

import (
	"bytes"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const MAX = 100

type Request struct {
	FileId bson.ObjectId
	SubId  bson.ObjectId
}

type Source struct {
	Id      bson.ObjectId
	Name    string
	Package string
	Ext     string
	Dir     string
}

func (s *Source) AbsPath() string {
	return filepath.Join(s.Dir, s.Package, s.FullName())
}

func (s *Source) FullName() string {
	return s.Name + "." + s.Ext
}

func (s *Source) ClassName() string {
	return s.Package + "." + s.Name
}

type TestBuilder struct {
	Tests   map[string]bool
	m       *sync.Mutex
	TestDir string
}

func NewTestBuilder() *TestBuilder {
	dir := filepath.Join(os.TempDir(), "tests")
	return &TestBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

func (t *TestBuilder) Setup(project string) (err error) {
	t.m.Lock()
	if !t.Tests[project] {
		tests, err := db.GetTests(project)
		if err == nil {
			err = utils.Unzip(filepath.Join(t.TestDir, project), tests.Data)
			if err == nil {
				t.Tests[project] = true
			}
		}
	}
	t.m.Unlock()
	return err
}

func RunTests(src *Source) (err error) {
	//Hardcode for now
	testdir := filepath.Join(os.TempDir(), "tests")
	cp := src.Dir + ":" + testdir
	tests := []*Source{&Source{src.Id, "EasyTests", "testing", "java", testdir}, &Source{src.Id, "AllTests", "testing", "java", testdir}}
	for _, test := range tests {
		stderr, stdout, err := RunCommand("javac", "-cp", cp, "-d", src.Dir, "-s", src.Dir, "-implicit:class", test.AbsPath())
		db.AddResults(src.Id, test.Name+"_compile_error", stderr.Bytes())
		db.AddResults(src.Id, test.Name+"_compile_warning", stdout.Bytes())
		if err == nil {
			stderr, stdout, err = RunCommand("java", "-cp", cp, "org.junit.runner.JUnitCore", test.ClassName()) //
			//fmt.Println("java", "-cp", cp,  "org.junit.runner.JUnitCore", test.ClassName())
			db.AddResults(src.Id, test.Name+"_run_error", stderr.Bytes())
			db.AddResults(src.Id, test.Name+"_run_result", stdout.Bytes())
		}
	}
	return err
}

func RunFB(src *Source) (err error) {
	fb := "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar"
	stderr, stdout, err := RunCommand("java", "-jar", fb, "-textui", "-low", filepath.Join(src.Dir, src.Package))
	db.AddResults(src.Id, "findbugs_warnings", stderr.Bytes())
	db.AddResults(src.Id, "findbugs_warning_count", stdout.Bytes())
	return err
}

func RunTools(src *Source) {
	RunFB(src)
}

func setupSource(sourceId bson.ObjectId) (src *Source, err error) {
	f, err := db.GetFile(sourceId)
	if err == nil && f.IsSource() {
		//Specific to how the file names are formatted currently, should change.
		params := strings.Split(f.Name, "_")
		fname := strings.Split(params[len(params)-4], ".")
		pkg := params[len(params)-5]
		dir := filepath.Join(os.TempDir(), sourceId.Hex())
		src = &Source{sourceId, fname[0], pkg, fname[1], dir}
		err = utils.SaveFile(filepath.Join(dir, pkg), src.FullName(), f.Data)
	}
	return src, err
}

func setupTests(subId bson.ObjectId) (err error) {
	sub, err := db.GetSubmission(subId)
	if err == nil {
		err = testBuilder.Setup(sub.Project)
	}
	return err
}

var testBuilder *TestBuilder

func Serve(requests chan *Request) {
	testBuilder = NewTestBuilder()
	// Start handlers
	for r := range requests {
		go Process(r)
	}
	utils.Log("completed")
}

func Process(r *Request) {
	src, err := setupSource(r.FileId)
	if err == nil && src != nil {
		err = setupTests(r.SubId)
		if err == nil {
			err = RunTests(src)
			if err == nil {
				RunTools(src)
			}
		}
	}
	if err != nil {
		utils.Log(err)
	}
}

func RunCommand(args ...string) (stdout, stderr bytes.Buffer, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	return stdout, stderr, err
}
