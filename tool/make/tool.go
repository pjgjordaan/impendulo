package make

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	Tool struct {
		cmd  string
		path string
	}
)

const (
	NAME = "Make"
)

func New(mf *Makefile, dir string) (ret *Tool, err error) {
	cmd, err := config.MAKE.Path()
	if err != nil {
		return
	}
	makeInfo := tool.NewTarget("Makefile", "", dir, tool.C)
	err = util.SaveFile(makeInfo.FilePath(), mf.Data)
	if err != nil {
		return
	}
	ret = &Tool{
		cmd:  cmd,
		path: makeInfo.FilePath(),
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
	args := []string{this.cmd, "-C", ti.Dir, "-f", this.path}
	execRes := tool.RunCommand(args, nil)
	if execRes.Err != nil {
		if !tool.IsEndError(execRes.Err) {
			err = execRes.Err
		} else {
			//Unsuccessfull compile.
			res, err = gcc.NewResult(fileId, execRes.StdErr)
			if err == nil {
				err = tool.NewCompileError(ti.FullName(), string(execRes.StdErr))
			}
		}
	} else if execRes.HasStdErr() {
		//Compiler warnings.
		res, err = gcc.NewResult(fileId, execRes.StdErr)
	} else {
		res, err = gcc.NewResult(fileId, tool.COMPILE_SUCCESS)
	}
	return

}
