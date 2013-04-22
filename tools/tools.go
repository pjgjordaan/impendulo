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
)

const (
	DIR_PATH = iota
	PKG_PATH
	FILE_PATH
)

var FB *Tool

func init() {
	FB = &Tool{"findbugs", "java", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warnings", "warning_count", []string{"java", "-jar"}, []string{"-textui", "-low"}, PKG_PATH}
}

type TargetInfo struct {
	Name    string
	Lang string
	Package string
	Ext     string
	Dir     string
}

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

type Tool struct {
	Name     string   "_id"
	Lang string "lang"
	Exec     string   "exec"
	ErrName  string   "err"
	OutName  string   "out"
	Preamble []string "pre"
	Flags    []string "flags"
	Target   int      "target"
}

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

func ReadTool(tmap bson.M) *Tool {
	name := tmap["_id"].(string)
	lang := tmap["lang"].(string)
	exec := tmap["exec"].(string)
	errName := tmap["err"].(string)
	outName := tmap["out"].(string)
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
	return &Tool{name, lang, exec, errName, outName, pre, flags, target}
}

func RunTool(id bson.ObjectId, ti *TargetInfo, tool *Tool) {
	args := tool.GetArgs(ti.GetTarget(tool.Target))
	stderr, stdout, err := RunCommand(args...)
	if err != nil {
		utils.Log(stderr.String(), stdout.String(), err)
	}
	AddResults(id, tool.Name, tool.ErrName, stderr.Bytes())
	AddResults(id, tool.Name, tool.OutName, stdout.Bytes())
}

func Compile(id bson.ObjectId, ti *TargetInfo) bool {
	//Hardcode for now
	stderr, stdout, err := RunCommand("javac", "-implicit:class", ti.FilePath())
	AddResults(id, "compile", "error", stderr.Bytes())
	AddResults(id, "compile", "warning", stdout.Bytes())
	if err != nil {
		utils.Log(stderr.String(), stdout.String(), err)
	}
	return err == nil
}

func RunTests(id bson.ObjectId, project string, ti *TargetInfo) {
	testdir := filepath.Join(os.TempDir(), "tests", project)
	cp := ti.Dir + ":" + testdir
	//Hardcode for now
	tests := []*TargetInfo{&TargetInfo{"EasyTests", "java", "testing", "java", testdir}, &TargetInfo{"AllTests","java", "testing", "java", testdir}}
	for _, test := range tests {
		stderr, stdout, err := RunCommand("javac", "-cp", cp, "-d", ti.Dir, "-s", ti.Dir, "-implicit:class", test.FilePath())
		AddResults(id, test.Name, "compile_error", stderr.Bytes())
		AddResults(id, test.Name, "compile_warning", stdout.Bytes())
		if err != nil {
			utils.Log(stderr.String(), stdout.String(), err)
		}
		//compiled successfully
		if err == nil {
			stderr, stdout, err = RunCommand("java", "-cp", cp, "org.junit.runner.JUnitCore", test.Executable()) //
			AddResults(id, test.Name, "run_error", stderr.Bytes())
			AddResults(id, test.Name, "run_result", stdout.Bytes())
			if err != nil {
				utils.Log(stderr.String(), stdout.String(), err)
			}
		}
	}
}

func RunTools(id bson.ObjectId, ti *TargetInfo) {
	tools, err := db.GetAll(db.TOOLS, bson.M{"lang": ti.Lang})
	//db error
	if err != nil {
		panic(err)
	}
	for _, t := range tools {
		tool := ReadTool(t)
		RunTool(id, ti, tool)
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

func AddResults(fileId bson.ObjectId, name, result string, data []byte) {
	matcher := bson.M{"_id": fileId}
	change := bson.M{"$push": bson.M{"results": bson.M{name: bson.M{"result": result, "data": data}}}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		panic(err)
	}
}

func AddTool(tool *Tool) {
	err := db.UpsertId(db.TOOLS, tool.Name, tool)
	if err != nil {
		panic(err)
	}
}
