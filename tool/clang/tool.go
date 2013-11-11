package gcc

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type (
	Tool struct {
		cmd  string
		path string
	}
)

const (
	NAME = "Clang"
)

func New() (ret *Tool, err error) {
	cmd, err := config.CLANG.Path()
	if err != nil {
		return
	}
	ret = &Tool{
		cmd: cmd,
	}
	return
}

//Lang
func (this *Tool) Lang() tool.Language {
	return tool.C
}

//Name
func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	args := []string{this.cmd, "-Wall", "-Wextra", "-Wno-variadic-macros", "-pedantic", "-O0", "-o", ti.Name, ti.FilePath()}
	execRes := tool.RunCommand(args, nil)
	if execRes.Err != nil {
		if !tool.IsEndError(execRes.Err) {
			err = execRes.Err
		} else {
			//Unsuccessfull compile.
			res, err = NewResult(fileId, execRes.StdErr)
			if err == nil {
				err = tool.NewCompileError(ti.FullName(), string(execRes.StdErr))
			}
		}
	} else if execRes.HasStdErr() {
		//Compiler warnings.
		res, err = NewResult(fileId, execRes.StdErr)
	} else {
		res, err = NewResult(fileId, tool.COMPILE_SUCCESS)
	}
	return

}
