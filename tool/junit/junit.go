package junit

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"labix.org/v2/mgo/bson"
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
	return tool.JUNIT
}

func (this *JUnit) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (tool.Result, error) {
	comp := javac.NewJavac(this.cp)
	_, err := comp.Run(fileId, ti)
	if err != nil {
		return nil, err
	}
	target := ti.GetTarget(tool.EXEC_PATH)
	args := []string{this.java, "-cp", this.cp, "-Ddata.location=" + this.datalocation, this.exec, target}
	stdout, stderr, err := tool.RunCommand(args)
	if err != nil {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0 {
		return NewResult(fileId, ti.Name, stderr), nil
	}
	return NewResult(fileId, ti.Name, stdout), nil
}
