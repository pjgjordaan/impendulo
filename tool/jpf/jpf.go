package jpf

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"strings"
)

type JPF struct {
	java         string
	compCP          string
	execCP string
	file file *project.JPFFile
}

func NewJPF(file *project.JPFFile) *JPF {
	compCP := config.GetConfig(config.JPF_JAR)+":"+config.GetConfig(config.GSON_JAR)
	execCP := config.GetConfig(config.RUNJPF_JAR)+":"+config.GetConfig(config.GSON_JAR)
	return &JPF{config.GetConfig(config.JAVA), compCP, execCP, file}
}

func (this *JPF) GetLang() string {
	return "java"
}

func (this *JPF) GetName() string {
	return tool.JPF
}

func (this *JPF) GetArgs(target string) []string {
	return []string{}
}

func (this *JPF) compileArgs(jpfInfo *tool.TargetInfo){
	return []string{this.java, "-cp", jpfInfo.Dir+":"+this.compCP, jpfInfo.FilePath()}
}


func (this *JPF) execArgs(jpfConfig string, jpfInfo, target *tool.TargetInfo){
	return []string{this.java, "-cp", jpfInfo.Dir+":"+this.execCP, jpfInfo.Executable(), jpfConfig, target.Executable(), target.Dir}
}

func (this *JPF) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (*tool.Result, error) {
	err := util.Copy(ti.Dir, config.GetConfig(config.RUNNER_DIR))
	if err != nil {
		return nil, err
	}
	err = util.SaveFile(ti.Dir, this.file.Name, this.file.Data)
	if err != nil {
		return nil, err
	}
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", ti.Dir)
	stderr, stdout, err := tool.RunCommand(this.compileArgs(jpfInfo)...)
	if err != nil {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0 {
		return nil, fmt.Errorf("Could not compile jpf runner: %q.", string(stderr))
	}
	stderr, stdout, err := tool.RunCommand(this.execArgs(jpfInfo)...)
	
	return tool.NewResult(fileId, this, stdout), nil
}

func (this *JPF) GenHTML() bool {
	return false
}
