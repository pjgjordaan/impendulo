package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

//Findbugs is a tool.Tool used to run the Findbugs static analysis tool on
//Java classes.
type FindBugs struct {
	cmd string
}

//Creates a new instance of the Findbugs tool.
func New() *FindBugs {
	return &FindBugs{config.Config(config.FINDBUGS)}
}

func (this *FindBugs) Lang() string {
	return "java"
}

func (this *FindBugs) Name() string {
	return NAME
}

func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "findbugs.xml")
	args := []string{config.Config(config.JAVA), "-jar", this.cmd, "-textui", "-effort:max",
		"-experimental", "-xml:withMessages", "-relaxed", "-output", outFile, ti.PackagePath()}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run findbugs: %q.",
			string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
