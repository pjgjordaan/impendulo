package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"github.com/godfried/impendulo/util"
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
	outFile := filepath.Join(ti.Dir, "findbugs.xml")
	args := []string{config.GetConfig(config.JAVA), "-jar", this.cmd,
		"-textui", "-low", "-xml:withMessages", "-output", outFile, ti.PackagePath()}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil{
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
