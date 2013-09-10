package jpf

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

//JPF is a tool.Tool which runs a JPF on a Java source file.
type Tool struct {
	cp, jpfPath, exec string
}

//New creates a new JPF instance. jpfDir is the location of the
//Java JPF runner files. configPath is the location of the JPF
//configuration file.
func New(jpfConfig *Config, jpfDir string) (jpf *Tool, err error) {
	err = util.Copy(jpfDir, config.Config(config.JPF_RUNNER_DIR))
	if err != nil {
		return
	}
	jpfPath := filepath.Join(jpfDir, "config.jpf")
	err = util.SaveFile(jpfPath, jpfConfig.Data)
	if err != nil {
		return
	}
	cp := jpfDir + ":" + config.Config(config.JPF_JAR) + ":" +
		config.Config(config.RUNJPF_JAR) + ":" + config.Config(config.GSON_JAR)
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", jpfDir)
	pubInfo := tool.NewTarget("ImpenduloPublisher.java", "java", "runner", jpfDir)
	comp := javac.New(cp)
	id := bson.NewObjectId()
	_, err = comp.Run(id, jpfInfo)
	if err != nil {
		return
	}
	_, err = comp.Run(id, pubInfo)
	if err != nil {
		return
	}
	jpf = &Tool{
		cp:      cp,
		jpfPath: jpfPath,
		exec:    jpfInfo.Executable(),
	}
	return
}

func (this *Tool) Lang() string {
	return "java"
}

func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "jpf")
	args := []string{config.Config(config.JAVA), "-cp", ti.Dir + ":" +
		this.cp, this.exec, this.jpfPath, ti.Executable(),
		ti.Dir, outFile}
	outFile = outFile + ".xml"
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
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

func Allowed(key string) bool {
	_, ok := reserved[key]
	return !ok
}

type empty struct{}

var reserved = map[string]empty{
	"search.class":                  empty{},
	"listener":                      empty{},
	"target":                        empty{},
	"report.publisher":              empty{},
	"report.xml.class":              empty{},
	"report.xml.file":               empty{},
	"classpath":                     empty{},
	"report.xml.start":              empty{},
	"report.xml.transition":         empty{},
	"report.xml.constraint":         empty{},
	"report.xml.property_violation": empty{},
	"report.xml.show_steps":         empty{},
	"report.xml.show_method":        empty{},
	"report.xml.show_code":          empty{},
	"report.xml.finished":           empty{},
}
