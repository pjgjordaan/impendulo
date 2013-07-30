package checkstyle

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type Checkstyle struct {
	java       string
	cmd        string
	configFile string
}

func NewCheckstyle() *Checkstyle {
	return &Checkstyle{config.GetConfig(config.JAVA), config.GetConfig(config.CHECKSTYLE), config.GetConfig(config.CHECKSTYLE_CONFIG)}
}

func (this *Checkstyle) GetLang() string {
	return "java"
}

func (this *Checkstyle) GetName() string {
	return NAME
}

func (this *Checkstyle) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	args := []string{this.java, "-jar", this.cmd, "-f", "xml", "-c", this.configFile, "-r", ti.Dir}
	execRes := tool.RunCommand(args, nil)
	if execRes.HasStdOut() {
		res, err = NewResult(fileId, execRes.StdOut)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run checkstyle: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
