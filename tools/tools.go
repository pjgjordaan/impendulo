package tools

import (
	"bytes"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//Tool configuration for testing
func init() {
	_, err := db.GetOne(db.TOOLS, bson.M{NAME: "findbugs"})
	if err != nil {
		fb := &Tool{bson.NewObjectId(), "findbugs", "java", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", "warnings", []string{"java", "-jar"}, []string{"-textui", "-low"}, bson.M{"":""}, PKG_PATH}
		AddTool(fb)
	}
	_, err = db.GetOne(db.TOOLS, bson.M{NAME: COMPILE})
	if err != nil {
		javac := &Tool{bson.NewObjectId(), COMPILE, "java", "javac", "warnings", "errors", []string{""}, []string{"-implicit:class"}, bson.M{CP:""}, FILE_PATH}
		AddTool(javac)
	}
}

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
Used by tool to specify target 
*/
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
Retrieves a tool from a mongo map.
*/
func ReadTool(tmap bson.M) (*Tool,error) {
	id, err := utils.GetID(tmap, ID)
	if err != nil {
		return nil, err
	}
	name, err := utils.GetString(tmap, NAME)
	if err != nil {
		return nil, err
	}
	lang, err := utils.GetString(tmap, LANG)
	if err != nil {
		return nil, err
	}
	exec, err := utils.GetString(tmap, EXEC)
	if err != nil {
		return nil, err
	}
	outName, err := utils.GetString(tmap, OUT)
	if err != nil {
		return nil, err
	}
	errName, err := utils.GetString(tmap, ERR)
	if err != nil {
		return nil, err
	}
	pre, err := utils.GetStrings(tmap, PRE)
	if err != nil {
		return nil, err
	}
	flags, err := utils.GetStrings(tmap, FLAGS)
	if err != nil {
		return nil, err
	}
	argflags, err := utils.GetM(tmap, ARGFLAGS)
	if err != nil {
		return nil, err
	}
	target, err := utils.GetInt(tmap, TARGET)
	if err != nil {
		return nil, err
	}
	return &Tool{id, name, lang, exec, outName, errName, pre, flags, argflags, target}, err
}

/*
Runs a tool and records its results in the db.
*/
func RunTool(fileId bson.ObjectId, ti *TargetInfo, tool *Tool, fArgs map[string]string) (err error) {
	tool.setFlagArgs(fArgs)
	args := tool.GetArgs(ti.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
	AddResult(NewResult(fileId, tool.Id, tool.Name, tool.OutName, tool.ErrName, stdout.Bytes(), stderr.Bytes()))
	if err != nil {
		utils.Log(stderr.String(), stdout.String(), err)
	}
	return err
}

/*
Compiles a java source file and writes the results thereof to the database.
*/
func Compile(fileId bson.ObjectId, ti *TargetInfo) bool {
	cint, err := db.GetOne(db.TOOLS, bson.M{LANG: ti.Lang, NAME: COMPILE})
	if err != nil {
		utils.Log(ti.Lang+" compiler not found: ", err)
		return false
	}
	comp, err := ReadTool(cint)
	if err != nil {
		utils.Log("Could not read compiler: ", err)
		return false
	}
	err = RunTool(fileId, ti, comp, map[string]string{CP : ti.Dir})
	if err != nil {
		utils.Log("Unsuccesful compile: ", err)
		return false
	}
	return true
}

/*
Compiles a java source file and writes the results thereof to the database.
*/
func AlreadyCompiled(fileId bson.ObjectId, ti *TargetInfo) {
	cint, err := db.GetOne(db.TOOLS, bson.M{LANG: ti.Lang, NAME: COMPILE})
	if err != nil {
		utils.Log(ti.Lang+" compiler not found: ", err)
		return
	}
	compiler, err := ReadTool(cint)
	if err != nil {
		utils.Log("Could not read compiler: ", err)
		return
	}
	AddResult(NewResult(fileId, compiler.Id, compiler.Name, compiler.OutName, compiler.ErrName, []byte(""), []byte("")))
}

/*
Runs a java test suite on a source file. Records the results in the db. 
*/
func RunTest(fileId bson.ObjectId, ti *TargetInfo, testName string) {
	testdir := filepath.Join(os.TempDir(), "tests", ti.Project, "src")
	test := &TargetInfo{ti.Project, testName, "java", "testing", "java", testdir}
	cp := ti.Dir + ":" + testdir
	//Hardcode for now
	stderr, stdout, err := RunCommand("javac", CP, cp, "-d", ti.Dir, "-s", ti.Dir, "-implicit:class", test.FilePath())
	AddResult(NewResult(fileId, fileId, test.Name+"_compile", "warnings", "errors", stdout.Bytes(), stderr.Bytes()))
	if err != nil {
		utils.Log("Test compile error ", err)
	}
	//compiled successfully
	if err == nil {
		stderr, stdout, err = RunCommand("java", CP, cp+":"+testdir, "org.junit.runner.JUnitCore", test.Executable()) //
		AddResult(NewResult(fileId, fileId, test.Name+"_execute", "results", "errors", stdout.Bytes(), stderr.Bytes()))
		if err != nil {
			utils.Log("Test run error ", err)
		}
	}
}

/*
 Runs all available tools on a file, skipping the tools which have already been
 run.
*/
func RunTools(fileId bson.ObjectId, ti *TargetInfo, alreadyRun bson.M) {
	tools, err := db.GetAll(db.TOOLS, bson.M{"lang": ti.Lang})
	//db error
	if err != nil {
		utils.Log("Error retrieving tools: ", ti.Lang)
		return
	}
	for _, tmap := range tools {
		//Check if tool has already been run
		if _, ok := alreadyRun[tmap["name"].(string)]; !ok {
			tool, err := ReadTool(tmap)
			if err != nil {
				utils.Log("Could not read tool: ", err)
				continue
			}
			err = RunTool(fileId, ti, tool, map[string]string{})
			if err != nil {
				utils.Log(tool.Name, " gave error: ", err)
			}
		}
	}
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

func AddResult(res *Result) {
	matcher := bson.M{"_id": res.FileId}
	change := bson.M{"$set": bson.M{"results." + res.Name: res.Id}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		utils.Log("Error adding results:", change)
		return
	}
	err = db.AddOne(db.RESULTS, res)
	if err != nil {
		utils.Log("Error adding results:", res)
	}
}

func AddTool(tool *Tool) {
	err := db.AddOne(db.TOOLS, tool)
	if err != nil {
		utils.Log("Error adding tool:", tool)
	}
}
