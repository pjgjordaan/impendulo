package java

import(
	"labix.org/v2/mgo/bson"
	"github.com/godfried/cabanga/config"
	"github.com/godfried/cabanga/tool"
)
type Javac struct{
	cmd string
	cp string
}

func NewJavac(cp string) *Javac{
	return &Javac{config.GetConfig(config.JAVAC), cp}	
}

func (this *Javac) GetLang() string{
	return "java"
}

func (this *Javac) GetName()string{
	return "javac"
}

func (this *Javac) GetArgs(target string)[]string{
	return []string{this.cmd, "-cp", this.cp, "-implicit:class", target}
}

func (this *Javac) Run(fileId bson.ObjectId, ti *tool.TargetInfo)(*tool.Result, error){
	target := ti.GetTarget(tool.FILE_PATH)
	args := this.GetArgs(target)
	stderr, stdout, ok, err := tool.RunCommand(args...)
	if !ok {
		return nil, err
	}
	return tool.NewResult(fileId, this, stdout, stderr, err), nil
}
