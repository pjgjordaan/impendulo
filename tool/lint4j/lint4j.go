package lint4j

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type Lint4j struct {
	cmd string
}

func NewLint4j() *Lint4j {
	return &Lint4j{config.GetConfig(config.LINT4J)}
}

func (this *Lint4j) GetLang() string {
	return "java"
}

func (this *Lint4j) GetName() string {
	return tool.LINT4J
}


func (this *Lint4j) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (tool.Result, error) {
	args := []string{this.cmd, "-v", "5", "-sourcepath", ti.Dir, ti.Executable()}
	stdout, stderr, err := tool.RunCommand(args...)
	fmt.Println(string(stdout), string(stderr))
	fmt.Println(args)
	if err != nil {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0 {
		return NewResult(fileId, stderr), nil
	}
	return NewResult(fileId, stdout), nil

}
