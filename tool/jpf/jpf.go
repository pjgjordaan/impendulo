package jpf

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"labix.org/v2/mgo/bson"
)

type JPF struct {
	cp      string
	jpfPath string
	jpfInfo *tool.TargetInfo
	pubInfo *tool.TargetInfo
}

func NewJPF(jpfDir, configPath string) *JPF {
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", jpfDir)
	pubInfo := tool.NewTarget("ImpenduloPublisher.java", "java", "runner", jpfDir)
	cp := jpfDir + ":" + config.GetConfig(config.JPF_JAR) + ":" + config.GetConfig(config.RUNJPF_JAR) + ":" + config.GetConfig(config.GSON_JAR)
	return &JPF{cp, configPath, jpfInfo, pubInfo}
}

func (this *JPF) GetLang() string {
	return "java"
}

func (this *JPF) GetName() string {
	return NAME
}

func (this *JPF) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	if this.jpfPath == ""{
		err = fmt.Errorf("No jpf configuration file available.")
		return
	}
	comp := javac.NewJavac(this.cp)
	_, err = comp.Run(fileId, this.pubInfo)
	if err != nil {
		return
	}
	_, err = comp.Run(fileId, this.jpfInfo)
	if err != nil {
		return
	}
	args := []string{config.GetConfig(config.JAVA), "-cp", ti.Dir + ":" + this.cp, this.jpfInfo.Executable(), this.jpfPath, ti.Executable(), ti.Dir}
	stdout, stderr, err := tool.RunCommand(args, nil)
	if stdout != nil {
		res, err = NewResult(fileId, stdout)
	} else if stderr != nil && len(stderr) > 0 {
		err = fmt.Errorf("Could not execute jpf runner: %q.", string(stderr))
	}
	return
}
