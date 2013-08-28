package jpf

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"os"
)

//JPF is a tool.Tool which runs a JPF on a Java source file.
type JPF struct {
	cp      string
	jpfPath string
	jpfInfo *tool.TargetInfo
}

//New creates a new JPF instance. jpfDir is the location of the
//Java JPF runner files. configPath is the location of the JPF
//configuration file.
func New(jpfFile *project.JPFFile, jpfDir string) (jpf *JPF, err error) {
	err = util.Copy(jpfDir, config.GetConfig(config.JPF_RUNNER_DIR))
	if err != nil {
		return
	}
	jpfPath := filepath.Join(jpfDir, jpfFile.Name)
	err = util.SaveFile(jpfPath, jpfFile.Data)
	if err != nil{
		return
	}
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", jpfDir)
	cp := jpfDir + ":" + config.GetConfig(config.JPF_JAR) + ":" +
		config.GetConfig(config.RUNJPF_JAR) + ":" + config.GetConfig(config.GSON_JAR)
	jpf = &JPF{
		cp:cp, 
		jpfPath: jpfPath, 
		jpfInfo: jpfInfo,
	}
	return
}

func (this *JPF) GetLang() string {
	return "java"
}

func (this *JPF) GetName() string {
	return NAME
}

func (this *JPF) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	if this.jpfPath == "" {
		err = fmt.Errorf("No jpf configuration file available.")
		return
	}
	comp := javac.New(this.cp)
	_, err = comp.Run(fileId, this.jpfInfo)
	if err != nil {
		return
	}
	outFile := filepath.Join(ti.Dir, "jpf")
	args := []string{config.GetConfig(config.JAVA), "-cp", ti.Dir + ":" +
		this.cp, this.jpfInfo.Executable(), this.jpfPath, ti.Executable(), 
		ti.Dir, outFile}
	outFile = outFile+".xml"
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil{
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not execute jpf runner: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
