package javac

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

//Javac is a tool.Tool used to compile Java source files.
type Javac struct {
	cmd string
	cp  string
}

//New creates a new Javac instance. cp is the classpath used when compiling.
func New(cp string) *Javac {
	return &Javac{config.GetConfig(config.JAVAC), cp}
}

func (this *Javac) GetLang() string {
	return "java"
}

func (this *Javac) GetName() string {
	return NAME
}

func (this *Javac) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	args := []string{this.cmd, "-cp", this.cp + ":" + ti.Dir, 
		"-implicit:class", ti.FilePath()}
	//Compile the file.
	execRes := tool.RunCommand(args, nil)
	if execRes.HasStdErr() {
		//Unsuccessfull compile.
		res = NewResult(fileId, execRes.StdErr)
		err = &CompileError{ti.FullName(), string(execRes.StdErr)}
	} else if execRes.Err != nil {
		//Error occured when attempting to run tool.
		err = execRes.Err
	} else {
		if !execRes.HasStdOut() {
			//Successfull compile.
			execRes.StdOut = []byte("Compiled successfully")
		}
		res = NewResult(fileId, execRes.StdOut)
	}
	return
}

//CompileError is used to indicate that compilation failed.
type CompileError struct {
	name string
	msg  string
}

func (this *CompileError) Error() string {
	return fmt.Sprintf("Could not compile %q due to: %q.", this.name, this.msg)
}

//IsCompileError checks whether an error is a CompileError.
func IsCompileError(err error) (ok bool) {
	_, ok = err.(*CompileError)
	return
}
