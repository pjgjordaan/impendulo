package tool

import(
	"labix.org/v2/mgo/bson"
"reflect"
"os/exec"
"fmt"
"bytes"
	"github.com/godfried/cabanga/config"

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

func (this *GenericTool) addArgs(args map[string]string) {
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

type Javac struct{
	cmd string
	cp string
}

func NewJavac(cp string) *Javac{
	return &Javac{config.GetConfig(config.JAVAC), cp}	
}

func (this *Javac) GetLang() string{
	return "java"
}

func (this *Javac) GetName()string{
	return "javac"
}

func (this *Javac) GetArgs(target string)[]string{
	return []string{this.cmd, "-cp", this.cp, "-implicit:class", target}
}

func (this *Javac) Run(fileId bson.ObjectId, ti *TargetInfo)(*Result, error){
	target := ti.GetTarget(FILE_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	return NewResult(fileId, this, stdout, stderr, err), nil
}

type JUnit struct{
	jar string
	exec string
	cp string
	datalocation string
}

func NewJUnit(cp, datalocation string) *JUnit{
	return &JUnit{config.GetConfig(config.JUNIT_JAR), config.GetConfig(config.JUNIT_EXEC), cp, datalocation}	
}

func (this *JUnit) GetLang() string{
	return "java"
}


func (this *JUnit) GetName()string{
	return "junit"
}

func (this *JUnit) GetArgs(target string)[]string{
	return []string{this.jar, "-cp", this.cp, "-Ddata.location="+this.datalocation, this.exec, target}
}

func (this *JUnit) Run(fileId bson.ObjectId, ti *TargetInfo)(*Result, error){
	target := ti.GetTarget(EXEC_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := RunCommand(args...)
	if !ok {
		return nil, err
	}
	return NewResult(fileId, this, stdout, stderr, err), nil
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

