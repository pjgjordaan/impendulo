package pmd

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"strings"
	"os"
	"path/filepath"
)

type PMD struct {
	cmd string
	rules string
}

func New(rules []string) *PMD {
	return &PMD{config.GetConfig(config.PMD), strings.Join(rules, ",")}
}

func (this *PMD) GetLang() string {
	return "java"
}

func (this *PMD) GetName() string {
	return NAME
}

func GetRules()[]string{
	return []string{
		"java-basic", "java-braces", "java-clone", "java-codesize",
		"java-comments", "java-controversial", "java-design", "java-empty",
		"java-finalizers", "java-imports", "java-j2ee", "java-javabeans", 
		"java-junit", "java-logging-jakarta-commons", "java-logging-java",
		"java-migrating", "java-naming", "java-optimizations",
		"java-strictexception", "java-strings", "java-sunsecure", "java-typeresolution",
		"java-unnecessary", "java-unusedcode",
	}
}

func (this *PMD) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "pmd.xml")
	args := []string{this.cmd, config.PMD, "-f", "xml", "-stress",
		"-shortnames", "-R", this.rules, "-r", outFile, "-d", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil{
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
