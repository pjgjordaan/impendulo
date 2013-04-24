package tools

import (
	"bytes"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
"time"
)

const (
	DIR_PATH = iota
	PKG_PATH
	FILE_PATH
)

//Tool configuration for testing
func init() {
	_, err := db.GetOne(db.TOOLS, bson.M{"name": "findbugs"}) 
	if err != nil{
		fb := &Tool{bson.NewObjectId(),"findbugs", "java", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", "warnings", []string{"java", "-jar"}, []string{"-textui", "-low"}, PKG_PATH}
		AddTool(fb)
	}
	_, err = db.GetOne(db.TOOLS, bson.M{"name": "compile"}) 
	if err != nil{
		javac := &Tool{bson.NewObjectId(), "compile", "java", "javac", "warnings", "errors", []string{}, []string{"-implicit:class"}, FILE_PATH} 
		AddTool(javac)
	}
}

/*
Information about the target file
*/
type TargetInfo struct {
	Name    string
	Lang string
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

func (ti *TargetInfo) Executable() string {
	return ti.Package + "." + ti.Name
}

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

func NewTarget(name, lang, pkg, dir string) *TargetInfo {
	split := strings.Split(name, ".")
	return &TargetInfo{split[0], lang, pkg, split[1], dir}
}

/*
Generic tool specification
*/
type Tool struct {
	Id bson.ObjectId "_id"
	Name     string   "name"
	Lang string "lang"
	Exec     string   "exec"
	OutName  string   "out"
	ErrName  string   "err"
	Preamble []string "pre"
	Flags    []string "flags"
	Target   int      "target"
}

/*
Setup tool arguments for execution.
*/
func (tool *Tool) GetArgs(target string) (args []string) {
	args = make([]string, len(tool.Preamble)+len(tool.Flags)+2)
	for i, p := range tool.Preamble {
		args[i] = p
	}
	args[len(tool.Preamble)] = tool.Exec
	start := len(tool.Preamble) + 1
	stop := start + len(tool.Flags)
	for j := start; j < stop; j++ {
		args[j] = tool.Flags[j-start]
	}
	args[stop] = target
	return args
}

/*
Retrieves a tool from a mongo map.
*/
func ReadTool(tmap bson.M) *Tool {
	id := tmap["_id"].(bson.ObjectId)
	name := tmap["name"].(string)
	lang := tmap["lang"].(string)
	exec := tmap["exec"].(string)
	outName := tmap["out"].(string)	
	errName := tmap["err"].(string)
	pint := tmap["pre"].([]interface{})
	pre := make([]string, len(pint))
	for i, pval := range pint {
		pre[i] = pval.(string)
	}
	fint := tmap["flags"].([]interface{})
	flags := make([]string, len(fint))
	for j, fval := range fint {
		flags[j] = fval.(string)
	}
	target := tmap["target"].(int)
	return &Tool{id, name, lang, exec, outName, errName, pre, flags, target}
}

/*
Runs a tool and writes its results to the database.
*/
func RunTool(fileId bson.ObjectId, ti *TargetInfo, tool *Tool) error {
	args := tool.GetArgs(ti.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
	if err != nil {
		utils.Log(stderr.String(), stdout.String(), err)
	}
	AddResult(NewResult(fileId, tool.Id, tool.Name, tool.OutName, tool.ErrName,stdout.Bytes(), stderr.Bytes()))
	return err
}

/*
Compiles a java source file and writes the results thereof to the database.
*/
func Compile(fileId bson.ObjectId, ti *TargetInfo) bool {
	cint, err := db.GetOne(db.TOOLS, bson.M{"lang": ti.Lang, "name": "compile"})
	if err != nil{
		utils.Log(ti.Lang+" compiler not found: ", err)
		return false
	}
	comp := ReadTool(cint)
	err = RunTool(fileId, ti, comp)
	if err != nil{
		utils.Log("Unsuccesful compile: ", err)
		return false
	}
	return true
}


/*
Runs java tests on a source file
*/
func RunTests(fileId bson.ObjectId, project string, ti *TargetInfo) {
	testdir := filepath.Join(os.TempDir(), "tests", project)
	cp := ti.Dir + ":" + testdir
	//Hardcode for now
	tests := []*TargetInfo{&TargetInfo{"EasyTests", "java", "testing", "java", testdir}, &TargetInfo{"AllTests","java", "testing", "java", testdir}}
	for _, test := range tests {
		stderr, stdout, err := RunCommand("javac", "-cp", cp, "-d", ti.Dir, "-s", ti.Dir, "-implicit:class", test.FilePath())
		AddResult(NewResult(fileId, fileId, test.Name, "compile_warning", "compile_error",stdout.Bytes(), stderr.Bytes()))
		if err != nil {
			utils.Log(stderr.String(), stdout.String(), err)
		}
		//compiled successfully
		if err == nil {
			stderr, stdout, err = RunCommand("java", "-cp", cp, "org.junit.runner.JUnitCore", test.Executable()) //
			AddResult(NewResult(fileId, fileId, test.Name, "run_result", "run_error",stdout.Bytes(), stderr.Bytes()))
			if err != nil {
				utils.Log(stderr.String(), stdout.String(), err)
			}
		}
	}
}

/*
Runs all available tools on a file.
*/
func RunTools(fileId bson.ObjectId, ti *TargetInfo) {
	tools, err := db.GetAll(db.TOOLS, bson.M{"lang": ti.Lang})
	//db error
	if err != nil {
		panic(err)
	}
	for _, t := range tools {
		tool := ReadTool(t)
		RunTool(fileId, ti, tool)
	}
}

/*
Executes a given external command.
*/
func RunCommand(args ...string) (stdout, stderr bytes.Buffer, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	return stdout, stderr, err
}

type Result struct{
	Id bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	ToolId bson.ObjectId "toolId"
	Name string "name"
	OutName string "outname"
	ErrName string "errname"
	OutData []byte "outdata"
	ErrData []byte "errdata"
	Time int64 "time"
}

func NewResult(fileId, toolId bson.ObjectId, name, outname,errname string, outdata, errdata []byte)*Result{
	return &Result{bson.NewObjectId(), fileId, toolId, name, outname, errname, outdata, errdata, time.Now().UnixNano()}
}
/*
Adds execution results to the database.
*/
func AddResult(res *Result) {
	matcher := bson.M{"_id": res.FileId}
	//change := bson.M{"$push": bson.M{"results": bson.M{name: bson.M{"result": result, "data": data}}}}
	change := bson.M{"$set": bson.M{"results."+res.Name: res.Id}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		panic(err)
	}
	err = db.AddOne(db.RESULTS, res)
	if err != nil {
		panic(err)
	}
}

/*
Adds a new tool to the database.
*/
func AddTool(tool *Tool) {
	err := db.AddOne(db.TOOLS, tool)
	if err != nil {
		panic(err)
	}
}
