package processing

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/java"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

//SetupTests extracts a project's tests from db to filesystem for execution.
//It creates and returns a new TestRunner.
func SetupTests(projectId bson.ObjectId, dir string) ([]*TestRunner, error) {
	proj, err := db.GetProject(bson.M{project.ID: projectId}, nil)
	if err != nil {
		return nil, err
	}
	tests, err := db.GetTests(bson.M{project.PROJECT_ID: projectId}, nil)
	if err != nil {
		return nil, err
	}
	ret := make([]*TestRunner, len(tests))
	for i, test := range tests {
		ret[i] = &TestRunner{tool.NewTarget(test.Name, proj.Lang, test.Package, dir)}
		err = util.SaveFile(ret[i].Info.PkgPath(), ret[i].Info.FullName(), test.Test)
		if err != nil {
			return nil, err
		}
		if len(test.Data) == 0 {
			continue
		}
		err = util.Unzip(dir, test.Data)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

//TestRunner is used to run tests on files compiled files.
type TestRunner struct {
	Info *tool.TargetInfo
}

//Execute sets up and runs tests on a compiled file.
func (this *TestRunner) Execute(f *project.File, dir string) error {
	compiled, err := this.Compile(f, dir)
	if err != nil {
		return err
	}
	if !compiled {
		return nil
	}
	return this.Run(f, dir)
}

//Compile compiles a test for the current file.
func (this *TestRunner) Compile(f *project.File, dir string) (bool, error) {
	cp := dir + ":" + this.Info.Dir + ":" + config.GetConfig(config.JUNIT_JAR)
	javac := java.NewJavac(cp)
	if _, ok := f.Results[this.Info.Name+"_"+javac.GetName()]; ok {
		return true, nil
	}
	res, err := javac.Run(f.Id, this.Info)
	if err != nil {
		return false, err
	}
	util.Log("Test compile result", res)
	res.Name = this.Info.Name + "_" + javac.GetName()
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return true, nil
}

//Run runs a test on the current file.
func (this *TestRunner) Run(f *project.File, dir string) error {
	ju := junit.NewJUnit(dir+":"+this.Info.Dir, this.Info.Dir)
	if _, ok := f.Results[this.Info.Name+"_"+ju.GetName()]; ok {
		return nil
	}
	res, err := ju.Run(f.Id, this.Info)
	util.Log("Test run result", res)
	if err != nil {
		return err
	}
	res.Name = this.Info.Name + "_" + ju.GetName()
	return AddResult(res)
}
