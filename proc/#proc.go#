package proc

import (
	"bytes"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/submission"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Source struct {
	Id      bson.ObjectId
	Name    string
	Package string
	Ext     string
	Dir     string
}

func (s *Source) FilePath() string {
	return filepath.Join(s.Dir, s.Package, s.FullName())
}

func (s *Source) PkgPath() string {
	return filepath.Join(s.Dir, s.Package)
}

func (s *Source) FullName() string {
	return s.Name + "." + s.Ext
}

func (s *Source) ClassName() string {
	return s.Package + "." + s.Name
}


func (src *Source) GetTarget(id int) (target string){
	switch id {
	case DIR_PATH: target = src.Dir
	case PKG_PATH: target = src.PkgPath()
	case FILE_PATH: target = src.FilePath()
	}
	return target
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
		tests, err := GetTests(project)
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

func GetTests(project string) (tests *submission.File, err error) {
	sint, err := db.GetMatching(db.SUBMISSIONS, bson.M{"project": project, "mode": "TEST"})
	if err == nil {
		sub := sint.(*submission.Submission)
		tint, err := db.GetMatching(db.FILES, bson.M{"subid": sub.Id})
		if err == nil{
			tests = tint.(*submission.File)
		}
	}
	return tests, err
}



func RunTests(src *Source) (err error) {
	//Hardcode for now
	testdir := filepath.Join(os.TempDir(), "tests")
	cp := src.Dir + ":" + testdir
	tests := []*Source{&Source{src.Id, "EasyTests", "testing", "java", testdir}, &Source{src.Id, "AllTests", "testing", "java", testdir}}
	for _, test := range tests {
		stderr, stdout, err := RunCommand("javac", "-cp", cp, "-d", src.Dir, "-s", src.Dir, "-implicit:class", test.FilePath())
		AddResults(src.Id, test.Name, "compile_error", stderr.Bytes())
		AddResults(src.Id, test.Name, "compile_warning", stdout.Bytes())
		if err == nil {
			stderr, stdout, err = RunCommand("java", "-cp", cp, "org.junit.runner.JUnitCore", test.ClassName()) //
			AddResults(src.Id, test.Name, "run_error", stderr.Bytes())
			AddResults(src.Id, test.Name,"run_result", stdout.Bytes())
		}
	}
	return err
}


func AddResults(fileId bson.ObjectId, name, result string,  data []byte)  error {
	matcher := bson.M{"_id": fileId}
	change := bson.M{"$push": bson.M{"results": bson.M{name: bson.M{"result":result, "data": data}}}}
	return db.Update(db.FILES, matcher, change)		
}


func RunFB(src *Source) (err error) {
	fb := "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar"
	stderr, stdout, err := RunCommand("java", "-jar", fb, "-textui", "-low", filepath.Join(src.Dir, src.Package))
	AddResults(src.Id, "findbugs", "warnings", stderr.Bytes())
	AddResults(src.Id, "findbugs", "warning_count", stdout.Bytes())
	return err
}

type Tool struct{
	Name string
	Exec string
	ErrName string
	OutName string
	Preamble []string
	Flags []string
	Target int
}

const(
DIR_PATH = iota
PKG_PATH
FILE_PATH
)

func (tool *Tool) GetArgs(target string) (args [] string){
	args = make([]string, len(tool.Preamble) + len(tool.Flags) + 2)
	for i, p := range tool.Preamble{
		args[i] = p
	}
	args[len(tool.Preamble)] = tool.Exec
	start := len(tool.Preamble)+1
	stop := start+len(tool.Flags)
	for j := start; j < stop; j++{
		args[j] = tool.Flags[j]
	}
	args[stop] = target
	return args
}


func RunTool(src *Source, tool *Tool)(err error){
	args := tool.GetArgs(src.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
	AddResults(src.Id, tool.Name, tool.ErrName, stderr.Bytes())
	AddResults(src.Id, tool.Name, tool.OutName, stdout.Bytes())
	return err
} 

func RunTools(src *Source) {
	tools,_ := db.GetAll(db.TOOLS)
	for _, t := range tools{
		tool, _ := t.(*Tool)
		RunTool(src, tool)
	}
}

func setupSource(f *submission.File) (src *Source, err error) {
	err = db.AddSingle(db.FILES, f) 
	if err == nil && f.IsSource() {
		dir := filepath.Join(os.TempDir(), f.Id.Hex())
		src = &Source{f.Id, fname[0], pkg, fname[1], dir}
		err = utils.SaveFile(filepath.Join(dir, pkg), src.FullName(), f.Data)
	}
	return src, err
}

func setupTests(subId bson.ObjectId) (err error) {
	sint, err := db.GetById(db.SUBMISSIONS, subId)
	if err == nil {
		sub := sint.(*submission.Submission)
		err = testBuilder.Setup(sub.Project)
	}
	return err
}

var testBuilder *TestBuilder

func Serve(files chan *submission.File) {
	testBuilder = NewTestBuilder()
	// Start handlers
	for f := range files {
		go Process(f)
	}
	utils.Log("completed")
}

func Process(f *submission.File) {
	src, err := setupSource(f)
	if err == nil && src != nil {
		err = setupTests(f.SubId)
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
