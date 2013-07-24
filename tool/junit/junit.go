package junit

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"labix.org/v2/mgo/bson"
	"fmt"
)

type JUnit struct {
	java         string
	exec         string
	cp           string
	datalocation string
}

func NewJUnit(cp, datalocation string) *JUnit {
	cp += ":" + config.GetConfig(config.JUNIT_JAR)
	return &JUnit{config.GetConfig(config.JAVA), config.GetConfig(config.JUNIT_EXEC), cp, datalocation}
}

func (this *JUnit) GetLang() string {
	return "java"
}

func (this *JUnit) GetName() string {
	return NAME
}

func (this *JUnit) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	comp := javac.NewJavac(this.cp)
	_, err = comp.Run(fileId, ti)
	if err != nil {
		return
	}
	args := []string{this.java, "-cp", this.cp, "-Ddata.location=" + this.datalocation, this.exec, ti.Executable()}
	stdout, stderr, err := tool.RunCommand(args, nil)
	if stdout != nil {
		res = NewResult(fileId, ti.Name, stdout)
		err = nil
	} else if stderr != nil && len(stderr) > 0 {
		err = fmt.Errorf("Could not run junit: %q.", string(stderr))
	}
	return
}
