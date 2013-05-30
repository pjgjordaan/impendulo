package junit

import(
	"labix.org/v2/mgo/bson"
	"github.com/godfried/cabanga/config"
	"github.com/godfried/cabanga/tool"
)


type JUnit struct{
	java string
	jar string
	exec string
	cp string
	datalocation string
}

func NewJUnit(cp, datalocation string) *JUnit{
	return &JUnit{config.GetConfig(config.JAVA), config.GetConfig(config.JUNIT_JAR), config.GetConfig(config.JUNIT_EXEC), cp, datalocation}	
}

func (this *JUnit) GetLang() string{
	return "java"
}


func (this *JUnit) GetName()string{
	return "junit"
}

func (this *JUnit) GetArgs(target string)[]string{
	return []string{this.java, "-jar", this.jar, "-cp", this.cp, "-Ddata.location="+this.datalocation, this.exec, target}
}

func (this *JUnit) Run(fileId bson.ObjectId, ti *tool.TargetInfo)(*tool.Result, error){
	target := ti.GetTarget(tool.EXEC_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := tool.RunCommand(args...)
	if !ok {
		return nil, err
	}
	return tool.NewResult(fileId, this, stdout, stderr, err), nil
}
