package tool

import (
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"strings"
	"time"
	"bytes"
	"os"
	"os/exec"
)

const (
	DIR_PATH = iota
	PKG_PATH
	FILE_PATH
	ID = "_id"
	NAME = "name"
	LANG = "lang"
	EXEC = "exec"
	OUT = "out"
	ERR = "err"
	PRE = "pre"
	FLAGS = "flags"
	ARGFLAGS = "argflags"
	TARGET = "target"
	CP = "-cp"
	COMPILE = "compile"
)


/*
Information about the target file
*/
type TargetInfo struct {
	Project string
	//File name without extension
	Name string
	//Language file is written in
	Lang    string
	Package string
	Ext     string
	Dir     string
}

/*
Helper functions to retrieve various target path strings used by tools.
*/
func (ti *TargetInfo) FilePath() string {
	return filepath.Join(ti.Dir, ti.Package, ti.FullName())
}

func (ti *TargetInfo) PkgPath() string {
	return filepath.Join(ti.Dir, ti.Package)
}

func (ti *TargetInfo) FullName() string {
	return ti.Name + "." + ti.Ext
}

/*
Path to compiled executable with package. 
*/
func (ti *TargetInfo) Executable() string {
	return ti.Package + "." + ti.Name
}

func (this *TargetInfo) GetCompiler()bson.M{
	return 	bson.M{LANG: this.Lang, NAME: COMPILE}
}

/*
Retrieves the target path based on the type required. 
*/
func (ti *TargetInfo) GetTarget(id int) (target string) {
	switch id {
	case DIR_PATH:
		target = ti.Dir
	case PKG_PATH:
		target = ti.PkgPath()
	case FILE_PATH:
		target = ti.FilePath()
	}
	return target
}

func NewTarget(project, name, lang, pkg, dir string) *TargetInfo {
	split := strings.Split(name, ".")
	return &TargetInfo{project, split[0], lang, pkg, split[1], dir}
}

/*
Generic tool specification
*/
type Tool struct {
	Id   bson.ObjectId "_id"
	Name string        "name"
	Lang string        "lang"
	//Tool executable, e.g. findbugs.jar
	Exec    string "exec"
	OutName string "out"
	ErrName string "err"
	//Arguments which occur prior to tool executable, e.g. java -jar
	Preamble []string "pre"
	//Flags which occur after executable, e.g. -v 
	Flags []string "flags"
	//Flags which occur after executable which require arguments, e.g. -cp src
	ArgFlags bson.M "argflags"
	//Tool target type, used to get the actual target. e.g. FILE_PATH
	Target int "target"
}

/*
Setup tool arguments for execution.
*/
func (this *Tool) GetArgs(target string) (args []string) {
	args = make([]string, len(this.Preamble)+len(this.Flags)+(len(this.ArgFlags)*2)+2)
	for i, p := range this.Preamble {
		args[i] = p
	}
	args[len(this.Preamble)] = this.Exec
	start := len(this.Preamble) + 1
	stop := start + len(this.Flags)
	for j := start; j < stop; j++ {
		args[j] = this.Flags[j-start]
	}
	cur := stop
	stop += len(this.ArgFlags)*2
	for k,v := range this.ArgFlags {
		args[cur] = k
		val := v.(string)
		args[cur + 1] = val
		cur += 2
	}
	args[stop] = target
	return args
}

func (this *Tool) setFlagArgs(args map[string] string){
	for k, arg := range args{
		if _, ok := this.ArgFlags[k]; ok{
			this.ArgFlags[k] = arg
		}
	}
	for flag, val := range this.ArgFlags{
		if strings.TrimSpace(val.(string)) == ""{
			delete(this.ArgFlags, flag)
		}
	}
}

/*
Describes a tool or test's results for a given file.
*/
type Result struct {
	Id      bson.ObjectId "_id"
	FileId  bson.ObjectId "fileid"
	ToolId  bson.ObjectId "toolId"
	Name    string        "name"
	OutName string        "outname"
	ErrName string        "errname"
	OutData []byte        "outdata"
	ErrData []byte        "errdata"
	Time    int64         "time"
}

func NewResult(fileId, toolId bson.ObjectId, name, outname, errname string, outdata, errdata []byte) *Result {
	return &Result{bson.NewObjectId(), fileId, toolId, name, outname, errname, outdata, errdata, time.Now().UnixNano()}
}


/*
Executes a given external command.
*/
func RunCommand(args ...string) (stdout, stderr bytes.Buffer, err error) {
	utils.Log(args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	return stdout, stderr, err
}

/*
Runs a tool and records its results in the db.
*/
func RunTool(fileId bson.ObjectId, ti *TargetInfo, tool *Tool, fArgs map[string]string) *Result {
	tool.setFlagArgs(fArgs)
	args := tool.GetArgs(ti.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
	if err != nil {
		utils.Log(stderr.String(), stdout.String(), err)
	}
	return NewResult(fileId, tool.Id, tool.Name, tool.OutName, tool.ErrName, stdout.Bytes(), stderr.Bytes())
}

func CompileTest(fileId bson.ObjectId, ti *TargetInfo, testName string)(*Result, error) {
	testdir := filepath.Join(os.TempDir(), "tests", ti.Project, "src")
	test := &TargetInfo{ti.Project, testName, "java", "testing", "java", testdir}
	cp := ti.Dir + ":" + testdir
	//Hardcode for now
	stderr, stdout, err := RunCommand("javac", CP, cp, "-d", ti.Dir, "-s", ti.Dir, "-implicit:class", test.FilePath())
	if err != nil {
		utils.Log("Test compile error ", err)
	}
	res := NewResult(fileId, fileId, test.Name+"_compile", "warnings", "errors", stdout.Bytes(), stderr.Bytes())
	return res, err
}

/*
Runs a java test suite on a source file. 
*/
func RunTest(fileId bson.ObjectId, ti *TargetInfo, testName string)(*Result, error) {
	testdir := filepath.Join(os.TempDir(), "tests", ti.Project, "src")
	test := &TargetInfo{ti.Project, testName, "java", "testing", "java", testdir}
	cp := ti.Dir + ":" + testdir
	stderr, stdout, err := RunCommand("java", CP, cp+":"+testdir, "org.junit.runner.JUnitCore", test.Executable())
	if err != nil {
		utils.Log("Test run error ", err)
	}
	res := NewResult(fileId, fileId, test.Name+"_compile", "warnings", "errors", stdout.Bytes(), stderr.Bytes())
	return res, err

}



