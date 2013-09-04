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
	j, err := JPF(proc)
	if err == nil {
		tools = append(tools, j)
	} else {
		util.Log(err)
	}
	p, err := PMD(proc)
	if err == nil {
		tools = append(tools, p)
	} else {
		util.Log(err)
	}
	tests, err := JUnit(proc)
	if err == nil {
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

func JPF(proc *Processor) (j tool.Tool, err error) {
	jpfFile, err := db.GetJPF(
		bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	j, err = jpf.New(jpfFile, proc.toolDir)
	return
}

func PMD(proc *Processor) (p tool.Tool, err error) {
	rules, err := db.GetPMD(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		rules = pmd.DefaultRules(proc.project.Id)
		err = db.AddPMD(rules)
	}
	p = pmd.New(rules.Rules)
	return
}

func JUnit(proc *Processor) (ret []tool.Tool, err error) {
	tests, err := db.GetTests(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	ret = make([]tool.Tool, len(tests))
	for i, test := range tests {
		ti := tool.NewTarget(test.Name, proc.project.Lang, test.Package, proc.toolDir)
		ret[i] = junit.New(ti)
		err = util.SaveFile(ti.FilePath(), test.Test)
		if err != nil {
			return
		}
		if len(test.Data) == 0 {
			continue
		}
		err = util.Unzip(ti.PackagePath(), test.Data)
		if err != nil {
			return
		}
	}
	err = util.Copy(proc.toolDir, config.Config(config.TESTING_DIR))
	return
}
