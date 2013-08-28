package checkstyle

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

type Checkstyle struct {
	java       string
	cmd        string
	configFile string
}

func New() *Checkstyle {
	return &Checkstyle{config.GetConfig(config.JAVA),
		config.GetConfig(config.CHECKSTYLE),
		config.GetConfig(config.CHECKSTYLE_CONFIG)}
}

func (this *Checkstyle) GetLang() string {
	return "java"
}

func (this *Checkstyle) GetName() string {
	return NAME
}

func (this *Checkstyle) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "checkstyle.xml")
	args := []string{this.java, "-jar", this.cmd,
		"-f", "xml", "-c", this.configFile, 
		"-o", outFile, "-r", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil{
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run checkstyle: %q.",
			string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
