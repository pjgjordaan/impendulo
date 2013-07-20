package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type FindBugs struct {
	cmd string
}

func NewFindBugs() *FindBugs {
	return &FindBugs{config.GetConfig(config.FINDBUGS)}
}

func (this *FindBugs) GetLang() string {
	return "java"
}

func (this *FindBugs) GetName() string {
	return NAME
}

func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	args :=  []string{config.GetConfig(config.JAVA), "-jar", this.cmd, "-textui", "-low", "-xml:withMessages", ti.PackagePath()}
	stdout, stderr, err := tool.RunCommand(args, nil)
	if stdout != nil {
		res, err = NewResult(fileId, stdout)
	} else if stderr != nil && len(stderr) > 0 {
		err = fmt.Errorf("Could not run findbugs: %q.", string(stderr))
	}
	return
}