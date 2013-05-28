package tool

import(
	"labix.org/v2/mgo/bson"
"reflect"
"strings"
"os/exec"
"fmt"
"bytes"
)

type Tool interface{
	func GetLang()string
	func Run(fileId bson.ObjectId, target *TargetInfo)(*Result, error)
}

//Tool is a generic tool specification.
type BasicTool struct{
	name string
	lang string
	exec string 
	outname string 
	errname string 
	preamble []string 
	flags []string 
	args bson.M 
	Target int 
}

//GetArgs sets up tool arguments for execution.
func (this *BasicTool) GetArgs(target string) (args []string) {
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

func (this *BasicTool) addArgs(args map[string]string) {
	this.ArgFlags = args
}

func (this *BasicTool) Equals(that Tool) bool {
	return reflect.DeepEqual(this, that)
}

func (this *BasicTool) Run(fileId bson.ObjectId, ti *TargetInfo)(*Result, error){
	target := ti.GetTarget(this.Target)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	return ToolResult(fileId, this, stdout, stderr, err), nil
}


type Javac struct{
	cmd string
	cp []string
}

func NewJavac(cmd string, cp []string) *Javac{
	return &Javac{cmd, cp}	
}

func (this *Javac) GetLang() string{
	return "java"
}

func (this *Javac) GetArgs(target string)[]string{
	return []string{this.cmd, "-cp", this.cp, "-implicit:class", target}
}

func Run(fileId bson.ObjectId, ti *TargetInfo)(*Result, error){
	target := ti.GetTarget(FILE_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	return ToolResult(fileId, this, stdout, stderr, err), nil
}

type JUnit


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

