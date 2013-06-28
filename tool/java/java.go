package java

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"strings"
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

func (this *Javac) args(target string) []string {
	return []string{this.cmd, "-cp", this.cp, "-implicit:class", target}
}

func (this *Javac) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (*tool.Result, error) {
	target := ti.GetTarget(tool.FILE_PATH)
	args := this.args(target)
	stderr, stdout, err := tool.RunCommand(args...)
	if err != nil {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0 {
		return tool.NewResult(fileId, this, stderr), nil
	}
	if stdout == nil || len(stdout) == 0 || len(strings.TrimSpace(string(stdout))) == 0 {
		stdout = []byte("Compiled successfully")
	}
	return tool.NewResult(fileId, this, stdout), nil
}

func (this *Javac) GenHTML() bool {
	return false
}
