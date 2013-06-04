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
	proj, err := db.GetProject(bson.M{project.ID: projectId})
	if err != nil {
		return nil, err
	}
	tests, err := db.GetTests(bson.M{project.PROJECT_ID: projectId})
	if err != nil {
		return nil, err
	}
	ret := make([]*TestRunner, len(tests))
	for i, test := range tests{
		testDir := filepath.Join(dir, test.Id.String())
		err = util.Unzip(testDir, test.Test)
		if err != nil {
			return nil, err
		}
		err = util.Unzip(testDir, test.Data)
		if err != nil {
			return nil, err
		}
		ret[i] = &TestRunner{test.Package, test.Name, proj.Lang, testDir}
	}
	return ret, nil
}


//TestRunner is used to run tests on files compiled files.
type TestRunner struct {
	Package string
	Name string
	Lang    string
	Dir     string
}

//Execute sets up and runs tests on a compiled file. 
func (this *TestRunner)  Execute(f *project.File, dir string) error {
	target := tool.NewTarget(this.Name, this.Lang, this.Package, this.Dir)
	compiled, err := this.Compile(target, f, dir)
	if err != nil {
		return err
	}
	if !compiled {
		return nil
	}
	return this.Run(target, f, dir)
}

//Compile compiles a test for the current file. 
func (this *TestRunner) Compile(target *tool.TargetInfo, f *project.File, dir string) (bool, error) {
	cp := dir+":"+this.Dir+":"+config.GetConfig(config.JUNIT_JAR)
	javac := java.NewJavac(cp)
	if _, ok := f.Results[target.Name+"_"+javac.GetName()]; ok {
		return true, nil
	}
	res, err := javac.Run(f.Id, target)
	if err != nil{
		return false, err
	}
	util.Log("Test compile result", res)
	res.Name = target.Name+"_"+javac.GetName()
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}




//Run runs a test on the current file.
func (this *TestRunner) Run(target *tool.TargetInfo, f *project.File, dir string) error {
	ju := junit.NewJUnit(dir+":"+this.Dir, this.Dir)
	if _, ok := f.Results[target.Name+"_"+ju.GetName()]; ok {
		return nil
	}
	res, err := ju.Run(f.Id, target)
	util.Log("Test run result", res)		
	if err != nil {
		return err
	}
	res.Name = target.Name+"_"+ju.GetName()
	return AddResult(res)
}
