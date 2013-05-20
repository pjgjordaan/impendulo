package tool

import (
	"bytes"
	"fmt"
	"labix.org/v2/mgo/bson"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"reflect"
)

//TargetInfo stores information about the target file.
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

//FilePath
func (ti *TargetInfo) FilePath() string {
	return filepath.Join(ti.Dir, ti.Package, ti.FullName())
}

//PkgPath
func (ti *TargetInfo) PkgPath() string {
	return filepath.Join(ti.Dir, ti.Package)
}

//FullName
func (ti *TargetInfo) FullName() string {
	return ti.Name + "." + ti.Ext
}

//Executable retrieves the path to the compiled executable with its package. 
func (ti *TargetInfo) Executable() string {
	return ti.Package + "." + ti.Name
}

//GetCompiler 
func (this *TargetInfo) GetCompiler() bson.M {
	return bson.M{LANG: this.Lang, NAME: COMPILE}
}

func (this *TargetInfo) Equals(that *TargetInfo) bool{
	return reflect.DeepEqual(this, that)
}

const (
	DIR_PATH = iota
	PKG_PATH
	FILE_PATH
)

//GetTarget retrieves the target path based on the type required. 
func (ti *TargetInfo) GetTarget(id int) string {
	switch id {
	case DIR_PATH:
		return ti.Dir
	case PKG_PATH:
		return ti.PkgPath()
	case FILE_PATH:
		return ti.FilePath()
	}
	return ""
}

//NewTarget
func NewTarget(project, name, lang, pkg, dir string) *TargetInfo {
	split := strings.Split(name, ".")
	return &TargetInfo{project, split[0], lang, pkg, split[1], dir}
}

//Tool is a generic tool specification.
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

//GetArgs sets up tool arguments for execution.
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
	stop += len(this.ArgFlags) * 2
	for k, v := range this.ArgFlags {
		args[cur] = k
		val := v.(string)
		args[cur+1] = val
		cur += 2
	}
	args[stop] = target
	return args
}

//setFlagArgs
func (this *Tool) setFlagArgs(args map[string]string) {
	for k, arg := range args {
		if _, ok := this.ArgFlags[k]; ok {
			this.ArgFlags[k] = arg
		}
	}
	for flag, val := range this.ArgFlags {
		if strings.TrimSpace(val.(string)) == "" {
			delete(this.ArgFlags, flag)
		}
	}
}

func (this *Tool) Equals(that *Tool) bool{
	return reflect.DeepEqual(this, that)
}

//Result describes a tool or test's results for a given file.
type Result struct {
	Id      bson.ObjectId "_id"
	FileId  bson.ObjectId "fileid"
	ToolId  bson.ObjectId "toolId"
	Name    string        "name"
	OutName string        "outname"
	ErrName string        "errname"
	OutData []byte        "outdata"
	ErrData []byte        "errdata"
	Error   error         "error"
	Time    int64         "time"
}

func (this *Result) Equals(that *Result) bool{
	return reflect.DeepEqual(this, that)
}


//NewResult
func NewResult(fileId, toolId bson.ObjectId, name, outname, errname string, outdata, errdata []byte, err error) *Result {
	return &Result{bson.NewObjectId(), fileId, toolId, name, outname, errname, outdata, errdata, err, time.Now().UnixNano()}
}

//RunCommand executes a given external command.
func RunCommand(args ...string) ([]byte, []byte, bool, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Start()
	if err != nil {
		return nil, nil, false, fmt.Errorf("Encountered error %q executing command %q", err, args)
	}
	err = cmd.Wait()
	return stdout.Bytes(), stderr.Bytes(), true, err
}

//RunTool runs a tool and records its results in the db.
func RunTool(fileId bson.ObjectId, ti *TargetInfo, tool *Tool, fArgs map[string]string) (*Result, error) {
	tool.setFlagArgs(fArgs)
	args := tool.GetArgs(ti.GetTarget(tool.Target))
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	res := NewResult(fileId, tool.Id, tool.Name, tool.OutName, tool.ErrName, stdout, stderr, err)
	return res, nil
}

//CompileTest compiles a java test suite against a given java source file.
func CompileTest(fileId bson.ObjectId, ti *TargetInfo, testName, testDir string) (*Result, error) {
	test := &TargetInfo{ti.Project, testName, JAVA, TESTS_PKG, JAVA, testDir}
	cp := ti.Dir + ":" + testDir
	stderr, stdout, ok, err := RunCommand(JAVAC, CP, cp, "-d", ti.Dir, "-s", ti.Dir, "-implicit:class", test.FilePath())
	if !ok {
		return nil, err
	}
	res := NewResult(fileId, fileId, test.Name+"_"+COMPILE, WARNS, ERRS, stdout, stderr, err)
	return res, nil
}

//RunTest runs a java test suite on a java class file. 
func RunTest(fileId bson.ObjectId, ti *TargetInfo, testName, testDir string) (*Result, error) {
	test := &TargetInfo{ti.Project, testName, JAVA, TESTS_PKG, JAVA, testDir}
	cp := ti.Dir + ":" + testDir
	stderr, stdout, ok, err := RunCommand(JAVA, CP, cp+":"+testDir, JUNIT, test.Executable())
	if !ok {
		return nil, err
	}
	res := NewResult(fileId, fileId, test.Name+"_"+RUN, WARNS, ERRS, stdout, stderr, err)
	return res, nil
}

const (
	CP        = "-cp"
	COMPILE   = "compile"
	RUN       = "run"
	NAME      = "name"
	LANG      = "lang"
	JUNIT     = "org.junit.runner.JUnitCore"
	JAVA      = "java"
	WARNS     = "warnings"
	ERRS      = "errors"
	TESTS_PKG = "testing"
	JAVAC     = "javac"
)
