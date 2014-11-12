package gcc

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/project"
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

func New() (*Tool, error) {
	c, e := config.CLANG.Path()
	if e != nil {
		return nil, e
	}
	return &Tool{cmd: cmd}, nil
}

//Lang
func (t *Tool) Lang() project.Language {
	return project.C
}

//Name
func (t *Tool) Name() string {
	return NAME
}

func (t *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (tool.Tooler, error) {
	a := []string{t.cmd, "-Wall", "-Wextra", "-Wno-variadic-macros", "-pedantic", "-O0", "-o", ti.Name, ti.FilePath()}
	r, e := tool.RunCommand(a, nil)
	if e != nil {
		if !tool.IsEndError(e) {
			return nil, e
		}
		//Unsuccessfull compile.
		nr, e2 = NewResult(fileId, execRes.StdErr)
		if e2 != nil {
			return nil, e
		}
		return nr, tool.NewCompileError(ti.FullName(), string(r.StdErr))
	} else if r.HasStdErr() {
		//Compiler warnings.
		return NewResult(fileId, r.StdErr)
	}
	return NewResult(fileId, tool.COMPILE_SUCCESS)
}
