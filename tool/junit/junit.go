package junit

import(
	"labix.org/v2/mgo/bson"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
)


type JUnit struct{
	java string
	exec string
	cp string
	datalocation string
}

func NewJUnit(cp, datalocation string) *JUnit{
	cp += ":"+config.GetConfig(config.JUNIT_JAR)
	return &JUnit{config.GetConfig(config.JAVA), config.GetConfig(config.JUNIT_EXEC), cp, datalocation}	
}

func (this *JUnit) GetLang() string{
	return "java"
}


func (this *JUnit) GetName()string{
	return tool.JUNIT
}

func (this *JUnit) GetArgs(target string)[]string{
	return []string{this.java, "-cp", this.cp, "-Ddata.location="+this.datalocation, this.exec, target}
}

func (this *JUnit) Run(fileId bson.ObjectId, ti *tool.TargetInfo)(*tool.Result, error){
	target := ti.GetTarget(tool.EXEC_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := tool.RunCommand(args...)
	if !ok {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0{
		return tool.NewResult(fileId, this, stderr), nil
	}
	return tool.NewResult(fileId, this, stdout), nil
}

func (this *JUnit) GenHTML() bool {
	return false
}