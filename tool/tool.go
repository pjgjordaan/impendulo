package tool
import(
	"labix.org/v2/mgo/bson"
"reflect"
"time"
"strings"
"path/filepath"
"os/exec"
"fmt"
"bytes"
)

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

func (this *Result) Equals(that *Result) bool {
	return reflect.DeepEqual(this, that)
}

//NewResult
func ToolResult(fileId bson.ObjectId, tool *Tool, outdata, errdata []byte, err error) *Result {
	return &Result{bson.NewObjectId(), fileId, tool.Id, tool.Name, tool.OutName, tool.ErrName, outdata, errdata, err, time.Now().UnixNano()}
}


//NewResult
func NewResult(fileId, toolId bson.ObjectId, name, outname, errname string, outdata, errdata []byte, err error) *Result {
	return &Result{bson.NewObjectId(), fileId, toolId, name, outname, errname, outdata, errdata, err, time.Now().UnixNano()}
}

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

func (this *TargetInfo) Equals(that *TargetInfo) bool {
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
type Tool struct{
	Id   bson.ObjectId "_id"
	Name string        "name"
	Lang string        "lang"
	Exec    string "exec"
	OutName string "out"
	ErrName string "err"
	Preamble []string "pre"
	Flags []string "flags"
	ArgFlags bson.M "argflags"
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

func (this *Tool) setArgFlags(args map[string]string) {
	if args != nil{
		for k, arg := range args {
			if _, ok := this.ArgFlags[k]; ok {
				this.ArgFlags[k] = arg
			}
		}
	}
	for flag, val := range this.ArgFlags {
		if strings.TrimSpace(val.(string)) == "" {
			delete(this.ArgFlags, flag)
		}
	}
}

func (this *Tool) Equals(that *Tool) bool {
	return reflect.DeepEqual(this, that)
}

func (this *Tool) Run(fileId bson.ObjectId, ti *TargetInfo, fArgs map[string]string)(*Result, error){
	this.setArgFlags(fArgs)
	target := ti.GetTarget(this.Target)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	return ToolResult(fileId, this, stdout, stderr, err), nil
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

