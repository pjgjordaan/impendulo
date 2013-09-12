package processing

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

func Tools(proc *Processor) (tools []tool.Tool, err error) {
	switch proc.project.Lang {
	case tool.JAVA:
		tools = javaTools(proc)
	default:
		err = fmt.Errorf("No tools found for %s language.",
			proc.project.Lang)
	}
	return
}

func javaTools(proc *Processor) []tool.Tool {
	tools := []tool.Tool{
		findbugs.New(),
		checkstyle.New(),
	}
	jpfTool, err := JPF(proc)
	if err == nil {
		tools = append(tools, jpfTool)
	} else {
		util.Log(err)
	}
	pmdTool, err := PMD(proc)
	if err == nil {
		tools = append(tools, pmdTool)
	} else {
		util.Log(err)
	}
	tests, err := JUnit(proc)
	if err == nil && len(tests) > 0 {
		tools = append(tools, tests...)
	} else {
		util.Log(err)
	}
	return tools
}

func Compiler(proc *Processor) (compiler tool.Tool, err error) {
	switch proc.project.Lang {
	case tool.JAVA:
		compiler = javac.New("")
	default:
		err = fmt.Errorf("No compiler found for %s language.",
			proc.project.Lang)
	}
	return
}

func JPF(proc *Processor) (runnable tool.Tool, err error) {
	jpfFile, err := db.GetJPF(
		bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	runnable, err = jpf.New(jpfFile, proc.toolDir)
	return
}

func PMD(proc *Processor) (runnable tool.Tool, err error) {
	rules, err := db.GetPMD(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		rules = pmd.DefaultRules(proc.project.Id)
		err = db.AddPMD(rules)
	}
	runnable = pmd.New(rules.Rules)
	return
}

func JUnit(proc *Processor) (ret []tool.Tool, err error) {
	tests, err := db.GetTests(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	err = util.Copy(proc.toolDir, config.Config(config.TESTING_DIR))
	if err != nil {
		return
	}
	ret = make([]tool.Tool, 0, len(tests))
	for _, test := range tests {
		unitTest, terr := junit.New(test, proc.toolDir)
		if terr != nil {
			util.Log(terr)
		} else {
			ret = append(ret, unitTest)
		}
	}
	return
}
