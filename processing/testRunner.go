package processing

import (
	"github.com/godfried/cabanga/config"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/tool/java"
	"github.com/godfried/cabanga/project"
	"github.com/godfried/cabanga/util"
	"github.com/godfried/cabanga/tool/junit"
	"labix.org/v2/mgo/bson"
	"path/filepath"
)

//SetupTests extracts a project's tests from db to filesystem for execution.
//It creates and returns a new TestRunner.
func SetupTests(projectId bson.ObjectId, dir string)([]*TestRunner, error) {
	proj, err := db.GetProject(bson.M{project.ID: projectId}, nil)
	if err != nil {
		return nil, err
	}
	tests, err := db.GetTests(bson.M{project.PROJECT_ID: projectId}, nil)
	if err != nil {
		return nil, err
	}
	ret := make([]*TestRunner, len(tests))
	for i, test := range tests{
		testDir := filepath.Join(dir, test.Id.String())
		ret[i] = &TestRunner{tool.NewTarget(test.Name, proj.Lang, test.Package, testDir)}
		err = util.SaveFile(filepath.Join(ret[i].Info.Dir, ret[i].Info.Package), ret[i].Info.FullName(), test.Test)
		if err != nil {
			return nil, err
		}
		err = util.Unzip(testDir, test.Data)
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
func (this *TestRunner)  Execute(f *project.File, dir string) error {
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
	cp := dir+":"+this.Info.Dir+":"+config.GetConfig(config.JUNIT_JAR)
	javac := java.NewJavac(cp)
	if _, ok := f.Results[this.Info.Name+"_"+javac.GetName()]; ok {
		return true, nil
	}
	res, err := javac.Run(f.Id, this.Info)
	if err != nil{
		return false, err
	}
	util.Log("Test compile result", res)
	res.Name = this.Info.Name+"_"+javac.GetName()
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
	res.Name = this.Info.Name+"_"+ju.GetName()
	return AddResult(res)
}
