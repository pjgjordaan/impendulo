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

type Tool struct {
	java       string
	cmd        string
	configFile string
}

func New() *Tool {
	return &Tool{config.Config(config.JAVA),
		config.Config(config.CHECKSTYLE),
		config.Config(config.CHECKSTYLE_CONFIG)}
}

func (this *Tool) Lang() string {
	return tool.JAVA
}

func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "checkstyle.xml")
	args := []string{this.java, "-jar", this.cmd,
		"-f", "xml", "-c", this.configFile,
		"-o", outFile, "-r", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
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
