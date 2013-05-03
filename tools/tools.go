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
	Project string
	//File name without extension
	Name    string
	//Language file is written in
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
)

/*
Generic tool specification
*/
type Tool struct {
	Id bson.ObjectId "_id"
	Name     string   "name"
	Lang string "lang"
	//Tool executable, e.g. findbugs.jar
	Exec     string   "exec"
	OutName  string   "out"
	ErrName  string   "err"
	//Arguments which occur prior to tool executable, e.g. java -jar
	Preamble []string "pre"
	//Flags which occur after executable, e.g. -v -cp
	Flags    []string "flags"
	//Tool target type, used to get the actual target. e.g. FILE_PATH
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
Runs a tool and records its results in the db.
*/
func RunTool(fileId bson.ObjectId, ti *TargetInfo, tool *Tool)(err error){
	args := tool.GetArgs(ti.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
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
Runs a java test suite on a source file. Records the results in the db. 
*/
func RunTest(fileId bson.ObjectId, ti *TargetInfo, testName string) {
	testdir := filepath.Join(os.TempDir(), "tests", ti.Project, "src")
	test := &TargetInfo{ti.Project, testName, "java", "testing", "java", testdir}
	cp := ti.Dir + ":" + testdir
	//Hardcode for now
	stderr, stdout, err := RunCommand("javac", "-cp", cp, "-d", ti.Dir, "-s", ti.Dir, "-implicit:class", test.FilePath())
	AddResult(NewResult(fileId, fileId, test.Name+"_compile", "warnings", "errors",stdout.Bytes(), stderr.Bytes()))
	if err != nil {
		utils.Log(stderr.String(), stdout.String(), err)
	}
	//compiled successfully
	if err == nil {
		stderr, stdout, err = RunCommand("java", "-cp", cp+":"+testdir, "org.junit.runner.JUnitCore", test.Executable()) //
		AddResult(NewResult(fileId, fileId, test.Name+"_execute", "results", "errors",stdout.Bytes(), stderr.Bytes()))
		if err != nil {
			utils.Log(stderr.String(), stdout.String(), err)
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
		if _, ok := alreadyRun[tmap["name"].(string)]; !ok{
			tool := ReadTool(tmap)
			err = RunTool(fileId, ti, tool)
			if err != nil{
				utils.Log(tool.Name, " gave error: ", err)
			}
		}
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

/*
Describes a tool or test's results for a given file.
*/
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

func AddResult(res *Result) {
	matcher := bson.M{"_id": res.FileId}
	change := bson.M{"$set": bson.M{"results."+res.Name: res.Id}}
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
