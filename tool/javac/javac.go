package javac

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
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
	return NAME
}

func (this *Javac) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	args := []string{this.cmd, "-cp", this.cp + ":" + ti.Dir, "-implicit:class", ti.FilePath()}
	execRes := tool.RunCommand(args, nil)
	if execRes.HasStdErr() {
		res = NewResult(fileId, execRes.StdErr)
		err = &CompileError{ti.FullName(), string(execRes.StdErr)}
	} else if execRes.Err != nil {
		err = execRes.Err
	} else {
		if !execRes.HasStdOut() {
			execRes.StdOut = []byte("Compiled successfully")
		}
		res = NewResult(fileId, execRes.StdOut)
	}
	return
}

type CompileError struct {
	name string
	msg  string
}

func (this *CompileError) Error() string {
	return fmt.Sprintf("Could not compile %q due to: %q.", this.name, this.msg)
}

func IsCompileError(err error) (ok bool) {
	_, ok = err.(*CompileError)
	return
}
