package junit

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

//Tool is a tool.Tool used to run Tool tests on a Java source file.
type Tool struct {
	cp           string
	datalocation string
	runnerInfo   *tool.TargetInfo
}

//New creates a new Tool instance. testDir is the location of the Tool testing files.
//cp is the classpath used and datalocation is the location of data files used when running
//the tests.
func New(testDir, cp, datalocation string) *Tool {
	runnerInfo := tool.NewTarget("TestRunner.java", "java", "testing", testDir)
	cp += ":" + config.GetConfig(config.JUNIT_JAR) + ":" +
		config.GetConfig(config.ANT_JUNIT) + ":" + config.GetConfig(config.ANT)
	return &Tool{
		cp: cp, 
		datalocation: datalocation, 
		runnerInfo: runnerInfo,
	}
}

func (this *Tool) GetLang() string {
	return "java"
}

func (this *Tool) GetName() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	//First compile the files to be tested
	comp := javac.New(this.cp)
	_, err = comp.Run(fileId, ti)
	if err != nil {
		return
	}
	//Compile the tests.
	_, err = comp.Run(fileId, this.runnerInfo)
	if err != nil {
		return
	}
	outFile := filepath.Join(this.datalocation, ti.Name+"_junit.xml")
	//Run the tests.
	args := []string{config.GetConfig(config.JAVA), "-cp", this.cp,
		this.runnerInfo.Executable(), ti.Executable(), this.datalocation}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil{
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, ti.Name, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run junit: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
