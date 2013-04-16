package proc

import (
	"bytes"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/sub"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"strings"
)


const(
DIR_PATH = iota
PKG_PATH
FILE_PATH
)

type SourceInfo struct {
	Name    string
	Package string
	Ext     string
	Dir     string
}

func (si *SourceInfo) FilePath() string {
	return filepath.Join(si.Dir, si.Package, si.FullName())
}

func (si *SourceInfo) PkgPath() string {
	return filepath.Join(si.Dir, si.Package)
}

func (si *SourceInfo) FullName() string {
	return si.Name + "." + si.Ext
}

func (si *SourceInfo) Executable() string {
	return si.Package + "." + si.Name
}


func (si *SourceInfo) GetTarget(id int) (target string){
	switch id {
	case DIR_PATH: target = si.Dir
	case PKG_PATH: target = si.PkgPath()
	case FILE_PATH: target = si.FilePath()
	}
	return target
}
func NewSourceInfo(name, pkg, dir string) *SourceInfo{
	split := strings.Split(name, ".")
	return &SourceInfo{split[0], pkg, split[1], dir}
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

func GetTests(project string) (tests *sub.File, err error) {
	smap, err := db.GetMatching(db.SUBMISSIONS, bson.M{"project": project, "mode": "TEST"})
	if err == nil {
		s := sub.ReadSubmission(smap)
		fmap, err := db.GetMatching(db.FILES, bson.M{"subid": s.Id})
		if err == nil{
			tests = sub.ReadFile(fmap)
		}
	}
	return tests, err
}



func RunTests(id bson.ObjectId, si *SourceInfo) (err error) {
	//Hardcode for now
	testdir := filepath.Join(os.TempDir(), "tests")
	cp := si.Dir + ":" + testdir
	tests := []*SourceInfo{&SourceInfo{"EasyTests", "testing", "java", testdir}, &SourceInfo{"AllTests", "testing", "java", testdir}}
	for _, test := range tests {
		stderr, stdout, err := RunCommand("javac", "-cp", cp, "-d", si.Dir, "-s", si.Dir, "-implicit:class", test.FilePath())
		AddResults(id, test.Name, "compile_error", stderr.Bytes())
		AddResults(id, test.Name, "compile_warning", stdout.Bytes())
		if err == nil {
			stderr, stdout, err = RunCommand("java", "-cp", cp, "org.junit.runner.JUnitCore", test.Executable()) //
			AddResults(id, test.Name, "run_error", stderr.Bytes())
			AddResults(id, test.Name,"run_result", stdout.Bytes())
		}
	}
	return err
}


func AddResults(fileId bson.ObjectId, name, result string,  data []byte)  error {
	matcher := bson.M{"_id": fileId}
	change := bson.M{"$push": bson.M{"results": bson.M{name: bson.M{"result":result, "data": data}}}}
	return db.Update(db.FILES, matcher, change)		
}


func RunFB(id bson.ObjectId, si *SourceInfo)(err error) {
	fb := "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar"
	stderr, stdout, err := RunCommand("java", "-jar", fb, "-textui", "-low", si.PkgPath())
	AddResults(id, "findbugs", "warnings", stderr.Bytes())
	AddResults(id, "findbugs", "warning_count", stdout.Bytes())
	return err
}

type Tool struct{
	Name string "name"
	Exec string "exec"
	ErrName string "err"
	OutName string "out"
	Preamble []string "pre"
	Flags []string "flags"
	Target int "target"
}

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

func ReadTool(tmap bson.M)*Tool{
	name := tmap["name"].(string)
	exec := tmap["exec"].(string)
	errName := tmap["err"].(string)
	outName := tmap["out"].(string)
	pre := tmap["pre"].([]string)
	flags := tmap["flags"].([]string)
	target := tmap["target"].(int)
	return &Tool{name, exec, errName, outName, pre, flags, target}
}

func RunTool(id bson.ObjectId, si *SourceInfo, tool *Tool)(err error){
	args := tool.GetArgs(si.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
	AddResults(id, tool.Name, tool.ErrName, stderr.Bytes())
	AddResults(id, tool.Name, tool.OutName, stdout.Bytes())
	return err
} 

func RunTools(id bson.ObjectId, si *SourceInfo) {
	tools,_ := db.GetAll(db.TOOLS)
	for _, t := range tools{
		tool := ReadTool(t)
		RunTool(id, si, tool)
	}
}


var testBuilder *TestBuilder

func Serve(files chan *sub.File) {
	testBuilder = NewTestBuilder()
	// Start handlers
	for f := range files {
		go Process(f)
	}
	utils.Log("completed")
}

func Process(f *sub.File) {
	t := f.Type()
	if t == sub.ARCHIVE{
		ProcessArchive(f)
	} else{
		err := db.AddSingle(db.FILES, f)
		if err == nil{
			if t == sub.SRC{
				ProcessSource(f)
			} else if t == sub.EXEC{
				ProcessExec(f)
			}else if t == sub.CHANGE{
				ProcessChange(f)
			}else if t == sub.TEST{
				ProcessTest(f)
			}
		}
	} 
}

func ProcessSource(src *sub.File){
	si, err := setupSource(src)
	if err == nil && si != nil {
		err = setupTests(src.SubId)
		if err == nil {
			err = RunTests(src.Id, si)
			if err == nil {
				RunTools(src.Id, si)
			}
		}
	}
	if err != nil {
		utils.Log(err)
	}
}


func setupSource(src *sub.File) (si *SourceInfo, err error) {
	dir := filepath.Join(os.TempDir(), src.Id.Hex())
	si  = NewSourceInfo(src.InfoStr(sub.NAME), src.InfoStr(sub.PKG), dir)
	err = utils.SaveFile(filepath.Join(dir, si.Package), si.FullName(), src.Data)
	return si, err
}

func setupTests(subId bson.ObjectId) (err error) {
	smap, err := db.GetById(db.SUBMISSIONS, subId)
	if err == nil {
		s := sub.ReadSubmission(smap)
		//sub := sint.(*sub.Submission)
		err = testBuilder.Setup(s.Project)
	}
	return err
}

func ProcessExec(s *sub.File){
}


func ProcessChange(s *sub.File){
}


func ProcessTest(s *sub.File){
}

func ProcessArchive(s *sub.File){
	files, err := utils.ReadZip(s.Data)
	if err == nil{
		for name, data := range files{
			info := sub.ParseName(name)
			f := sub.NewFile(s.SubId, info, data)
			Process(f)
		}
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
