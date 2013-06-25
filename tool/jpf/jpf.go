package jpf

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type JPF struct {
	java         string
	jar          string
	cp string
	jarexec bool
}

func NewJPF(cp string) *JUnit {
	return &JUnit{config.GetConfig(config.JAVA), config.GetConfig(config.JPF_JAR), cp, strings.HasSuffix(target, "jpf")}
}

func (this *JPF) GetLang() string {
	return "java"
}

func (this *JPF) GetName() string {
	return tool.JPF
}

func (this *JPF) GetArgs(target string) []string {
	if this.jarexec{
		return []string{this.java, "-cp", this.cp, "-jar", this.jar, target}
	} else{
		return []string{this.java, "-cp", this.cp+":"+this.jar, target}
	}

}

func (this *JPF) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (*tool.Result, error) {
	target := ti.GetTarget(tool.EXEC_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := tool.RunCommand(args...)
	if !ok {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0 {
		return tool.NewResult(fileId, this, stderr), nil
	}
	return tool.NewResult(fileId, this, stdout), nil
}

func (this *JPF) GenHTML() bool {
	return false
}
