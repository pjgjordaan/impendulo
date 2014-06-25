package gcc

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
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
	NAME = "GCC"
)

func New() (*Tool, error) {
	p, e := config.GCC.Path()
	if e != nil {
		return nil, e
	}
	return &Tool{cmd: p}, e
}

//Lang
func (t *Tool) Lang() tool.Language {
	return tool.C
}

//Name
func (t *Tool) Name() string {
	return NAME
}

func (t *Tool) AddCP(p string) {
}

func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (result.Tooler, error) {
	a := []string{t.cmd, "-Wall", "-Wextra", "-Wno-variadic-macros", "-pedantic", "-O0", "-o", target.Name, target.FilePath()}
	r, e := tool.RunCommand(a, nil, 30*time.Second)
	if e != nil {
		if !tool.IsEndError(e) {
			return nil, e
		}
		nr, e2 := NewResult(fileId, r.StdErr)
		if e2 != nil {
			return nil, e
		}
		return nr, tool.NewCompileError(target.FullName(), string(r.StdErr))
	} else if r.HasStdErr() {
		return NewResult(fileId, r.StdErr)
	}
	return NewResult(fileId, result.COMPILE_SUCCESS)
}
