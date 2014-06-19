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

func New(mf *Makefile, dir string) (*Tool, error) {
	cmd, e := config.MAKE.Path()
	if e != nil {
		return nil, e
	}
	t := tool.NewTarget("Makefile", "", dir, tool.C)
	if e = util.SaveFile(t.FilePath(), mf.Data); e != nil {
		return nil, e
	}
	return &Tool{
		cmd:  cmd,
		path: t.FilePath(),
	}, nil
}

func (t *Tool) AddCP(p string) {
}

//Lang
func (t *Tool) Lang() tool.Language {
	return tool.C
}

//Name
func (t *Tool) Name() string {
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
