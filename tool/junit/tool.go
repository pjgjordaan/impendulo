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
	cp                   string
	dataLocation         string
	testInfo, runnerInfo *tool.TargetInfo
}

//New creates a new Tool instance. testDir is the location of the Tool testing files.
//cp is the classpath used and datalocation is the location of data files used when running
//the tests.
func New(test *Test, dir string) (junit *Tool, err error) {
	testInfo := tool.NewTarget(test.Name, tool.JAVA, test.Package, dir)
	err = util.SaveFile(testInfo.FilePath(), test.Test)
	if err != nil {
		return
	}
	if len(test.Data) != 0 {
		err = util.Unzip(testInfo.PackagePath(), test.Data)
		if err != nil {
			return
		}
	}
	dataLocation := filepath.Join(testInfo.PackagePath(), "data")
	runnerInfo := tool.NewTarget("TestRunner.java", "java", "testing", testInfo.Dir)
	cp := testInfo.Dir + ":" + config.Config(config.JUNIT_JAR) + ":" +
		config.Config(config.ANT_JUNIT) + ":" + config.Config(config.ANT)
	junit = &Tool{
		cp:           cp,
		dataLocation: dataLocation,
		testInfo:     testInfo,
		runnerInfo:   runnerInfo,
	}
	return
}

func (this *Tool) Lang() string {
	return tool.JAVA
}

func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	if this.cp != "" {
		this.cp += ":"
	}
	this.cp += ti.Dir
	//First compile the files to be tested
	comp := javac.New(this.cp)
	_, err = comp.Run(fileId, this.testInfo)
	if err != nil {
		return
	}
	//Compile the tests.
	_, err = comp.Run(fileId, this.runnerInfo)
	if err != nil {
		return
	}
	outFile := filepath.Join(this.dataLocation, this.testInfo.Name+"_junit.xml")
	//Run the tests.
	args := []string{config.Config(config.JAVA), "-cp", this.cp,
		this.runnerInfo.Executable(), this.testInfo.Executable(), this.dataLocation}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, this.testInfo.Name, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run junit: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
