package junit

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"labix.org/v2/mgo/bson"
)

type JUnit struct {
	cp           string
	datalocation string
	runnerInfo *tool.TargetInfo
}

func NewJUnit(testDir, cp, datalocation string) *JUnit {
	runnerInfo := tool.NewTarget("TestRunner.java", "java", "testing", testDir)
	cp += ":" + config.GetConfig(config.JUNIT_JAR) + ":" + config.GetConfig(config.ANT_JUNIT) + ":" + config.GetConfig(config.ANT)
	return &JUnit{cp, datalocation, runnerInfo}
}

func (this *JUnit) GetLang() string {
	return "java"
}

func (this *JUnit) GetName() string {
	return NAME
}

func (this *JUnit) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	comp := javac.NewJavac(this.cp)
	_, err = comp.Run(fileId, ti)
	if err != nil {
		return
	}
	_, err = comp.Run(fileId, this.runnerInfo)
	if err != nil {
		return
	}
	args := []string{config.GetConfig(config.JAVA), "-cp", this.cp, this.runnerInfo.Executable(), ti.Executable(), this.datalocation}
	execRes := tool.RunCommand(args, nil)
	if execRes.HasStdOut() {
		res, err = NewResult(fileId, ti.Name, execRes.StdOut)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run junit: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
