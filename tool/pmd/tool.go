package pmd

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"strings"
)

type (
	//Tool
	Tool struct {
		cmd   string
		rules string
	}
)

//New
func New(rules *Rules) *Tool {
	if rules == nil {
		rules, _ = DefaultRules(bson.NewObjectId())
	}
	return &Tool{
		cmd:   config.Config(config.PMD),
		rules: strings.Join(rules.RuleArray(), ","),
	}
}

//Lang
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name
func (this *Tool) Name() string {
	return NAME
}

//Run
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "pmd.xml")
	args := []string{this.cmd, config.PMD, "-f", "xml", "-stress",
		"-shortnames", "-R", this.rules, "-r", outFile, "-d", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run pmd: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
