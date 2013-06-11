package findbugs

import(
	"labix.org/v2/mgo/bson"
	"github.com/godfried/cabanga/config"
	"github.com/godfried/cabanga/tool"
)
type FindBugs struct{
	cmd string
}

func NewFindBugs() *FindBugs{
	return &FindBugs{config.GetConfig(config.FINDBUGS)}	
}

func (this *FindBugs) GetLang() string{
	return "java"
}

func (this *FindBugs) GetName()string{
	return tool.FINDBUGS
}

func (this *FindBugs) GetArgs(target string)[]string{
	return []string{config.GetConfig(config.JAVA), "-jar", this.cmd, "-textui", "-low", "-html:fancy-hist.xsl", target}
}

func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo)(*tool.Result, error){
	target := ti.GetTarget(tool.PKG_PATH)
	args := this.GetArgs(target)
	stdout, stderr, ok, err := tool.RunCommand(args...)
	if !ok {
		return nil, err
	}
	return tool.NewResult(fileId, this, stdout, stderr, err), nil
}
