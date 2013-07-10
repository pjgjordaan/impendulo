package javac

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"strings"
	"fmt"
)

type Javac struct {
	cmd string
	cp  string
}

func NewJavac(cp string) *Javac {
	return &Javac{config.GetConfig(config.JAVAC), cp}
}

func (this *Javac) GetLang() string {
	return "java"
}

func (this *Javac) GetName() string {
	return tool.JAVAC
}

func (this *Javac) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (tool.Result, error) {
	target := ti.GetTarget(tool.FILE_PATH)
	args := []string{this.cmd, "-cp", this.cp+":"+ti.Dir, "-implicit:class", target}
	stdout, stderr, err := tool.RunCommand(args...)
	if stderr != nil && len(stderr) > 0 {
		return NewResult(fileId, stderr), &CompileError{ti.FullName(), string(stderr)}
	}else if err != nil{
		return nil, err
	}
	if stdout == nil || len(stdout) == 0 || len(strings.TrimSpace(string(stdout))) == 0 {
		stdout = []byte("Compiled successfully")
	}
	return NewResult(fileId, stdout), nil
}

type CompileError struct{
	name string
	msg string
}

func (this *CompileError) Error() string{
	return fmt.Sprintf("Could not compile %q due to: %q.", this.name, this.msg) 
}

func IsCompileError(err error)(ok bool){
	_, ok = err.(*CompileError)
	return
}