package tool

import(
	"labix.org/v2/mgo/bson"
"reflect"
"os/exec"
"fmt"
"bytes"
)

type Tool interface{
	GetName() string
	GetLang()string
	Run(fileId bson.ObjectId, target *TargetInfo)(*Result, error)
}

//Tool is a generic tool specification.
type GenericTool struct{
	name string
	lang string
	exec string 
	preamble []string 
	flags []string 
	args map[string]string
	target TargetSpec 
}

//GetArgs sets up tool arguments for execution.
func (this *GenericTool) GetArgs(target string) (args []string) {
	args = make([]string, len(this.preamble)+len(this.flags)+(len(this.args)*2)+2)
	for i, p := range this.preamble {
		args[i] = p
	}
	args[len(this.preamble)] = this.exec
	start := len(this.preamble) + 1
	stop := start + len(this.flags)
	for j := start; j < stop; j++ {
		args[j] = this.flags[j-start]
	}
	cur := stop
	stop += len(this.args) * 2
	for k, v := range this.args {
		args[cur] = k
		args[cur+1] = v
		cur += 2
	}
	args[stop] = target
	return args
}

func (this *GenericTool) AddArgs(args map[string]string) {
	this.args = args
}

func (this *GenericTool) Equals(that Tool) bool {
	return reflect.DeepEqual(this, that)
}

func (this *GenericTool) Run(fileId bson.ObjectId, ti *TargetInfo)(*Result, error){
	target := ti.GetTarget(this.target)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	return NewResult(fileId, this, stdout, stderr, err), nil
}

func (this *GenericTool) GetName()string{
	return this.name
}


func (this *GenericTool) GetLang()string{
	return this.lang
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

