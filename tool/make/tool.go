package make

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"time"
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

func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (tool.ToolResult, error) {
	r, e := tool.RunCommand([]string{t.cmd, "-C", target.Dir, "-f", t.path}, nil, 30*time.Second)
	if e != nil {
		if !tool.IsEndError(e) {
			return nil, e
		}
		//Unsuccessfull compile.
		nr, e := gcc.NewResult(fileId, r.StdErr)
		if e != nil {
			return nil, e
		}
		return nr, tool.NewCompileError(target.FullName(), string(r.StdErr))
	} else if r.HasStdErr() {
		//Compiler warnings.
		return gcc.NewResult(fileId, r.StdErr)
	}
	return gcc.NewResult(fileId, tool.COMPILE_SUCCESS)
}
