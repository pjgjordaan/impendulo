package jpf

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type JPF struct {
	exec    string
	cp  string
	jpfPath string
}

func NewJPF(jpfPath string) *JPF {
	cp := config.GetConfig(config.JPF_JAR) + ":" + config.GetConfig(config.RUNJPF_JAR) + ":" + config.GetConfig(config.GSON_JAR)
	return &JPF{config.GetConfig(config.JAVA), cp, jpfPath}
}

func (this *JPF) GetLang() string {
	return "java"
}

func (this *JPF) GetName() string {
	return tool.JPF
}

func (this *JPF) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	err = util.Copy(ti.Dir, config.GetConfig(config.RUNNER_DIR))
	if err != nil {
		return
	}
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", ti.Dir)
	comp := javac.NewJavac(this.cp)
	_, err = comp.Run(fileId, jpfInfo)
	if err != nil{
		return
	}
	args := []string{this.exec, "-cp", jpfInfo.Dir + ":" + this.cp, jpfInfo.Executable(), this.jpfPath, ti.Executable(), ti.Dir}
	stdout, stderr, err := tool.RunCommand(args...)
	if err == nil {
		if stderr != nil && len(stderr) > 0 {
			err = fmt.Errorf("Could not execute jpf runner: %q.", string(stderr))
		} else{
			res = NewResult(fileId, stdout)
		}
	}
	return
}
