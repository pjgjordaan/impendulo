package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

//Findbugs is a tool.Tool used to run the Findbugs static analysis tool on
//Java classes.
type FindBugs struct {
	cmd string
}

//Creates a new instance of the Findbugs tool.
func New() *FindBugs {
	return &FindBugs{config.GetConfig(config.FINDBUGS)}
}

func (this *FindBugs) GetLang() string {
	return "java"
}

func (this *FindBugs) GetName() string {
	return NAME
}

func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	args := []string{config.GetConfig(config.JAVA), "-jar", this.cmd,
		"-textui", "-low", "-xml:withMessages", ti.PackagePath()}
	execRes := tool.RunCommand(args, nil)
	if execRes.HasStdOut() {
		res, err = NewResult(fileId, execRes.StdOut)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run findbugs: %q.",
			string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
