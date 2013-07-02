package jpf

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"github.com/godfried/impendulo/util"
	"fmt"
)

type JPF struct {
	java         string
	javac string
	compCP          string
	execCP string
	jpfPath  string
}

func NewJPF(jpfPath  string) *JPF {
	compCP := config.GetConfig(config.JPF_JAR)+":"+config.GetConfig(config.GSON_JAR)
	execCP := config.GetConfig(config.JPF_JAR)+":"+config.GetConfig(config.RUNJPF_JAR)+":"+config.GetConfig(config.GSON_JAR)
	return &JPF{config.GetConfig(config.JAVA), config.GetConfig(config.JAVAC), compCP, execCP, jpfPath}
}

func (this *JPF) GetLang() string {
	return "java"
}

func (this *JPF) GetName() string {
	return tool.JPF
}

func (this *JPF) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (*tool.Result, error) {
	err := util.Copy(ti.Dir, config.GetConfig(config.IMP_JPF))
	if err != nil {
		return nil, err
	}
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", ti.Dir)
	compileArgs := []string{this.javac, "-cp", jpfInfo.Dir+":"+this.compCP, jpfInfo.FilePath()}
	_, stderr, err := tool.RunCommand(compileArgs...)
	if err != nil {
		return nil, err
	} else if stderr != nil && len(stderr) > 0 {
		return nil, fmt.Errorf("Could not compile jpf runner: %q.", string(stderr))
	}
	runArgs := []string{this.java, "-cp", jpfInfo.Dir+":"+this.execCP, jpfInfo.Executable(), this.jpfPath, ti.Executable(), ti.Dir}
	stdout, stderr, err := tool.RunCommand(runArgs...)	
	if err != nil {
		return nil, err
	} else if stderr != nil && len(stderr) > 0 {
		return nil, fmt.Errorf("Could not execute jpf runner: %q.", string(stderr))
	}
	return tool.NewResult(fileId, this, stdout), nil
}

func (this *JPF) GenHTML() bool {
	return false
}