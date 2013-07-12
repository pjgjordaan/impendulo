package pmd

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type PMD struct {
	cmd string
}

func NewPMD() *PMD {
	return &PMD{config.GetConfig(config.PMD)}
}

func (this *PMD) GetLang() string {
	return "java"
}

func (this *PMD) GetName() string {
	return NAME
}

const RULES = `java-basic,java-braces,java-clone,java-codesize,java-comments,java-controversial,java-design,java-empty,java-finalizers,java-imports,java-j2ee,java-javabeans,java-junit,java-logging-jakarta-commons,java-logging-java,java-migrating,java-naming,java-optimizations,java-strictexception,java-strings,java-sunsecure,java-typeresolution,java-unnecessary,java-unusedcode`

func (this *PMD) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	args := []string{this.cmd, config.PMD, "-f", "xml", "-stress", "-shortnames", "-rulesets", RULES, "-dir", ti.Dir}
	stdout, stderr, err := tool.RunCommand(args...)
	if stdout != nil {
		//var pres *PMDResult
		res, err = NewResult(fileId, stdout)
		//res = pres
	} else if stderr != nil && len(stderr) > 0 {
		err = fmt.Errorf("Could not run pmd: %q.", string(stderr))
	}
	return

}


