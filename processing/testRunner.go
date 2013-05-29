package processing

import (
	"github.com/godfried/cabanga/config"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
)

//SetupTests extracts a project's tests from db to filesystem for execution.
//It creates and returns a new TestRunner.
func SetupTests(project, lang, dir string) *TestRunner {
	testMatcher := bson.M{submission.PROJECT: project, submission.LANG: lang}
	test, err := db.GetTest(testMatcher)
	if err != nil {
		util.Log(err)
		return nil
	}
	err = util.Unzip(dir, test.Tests)
	if err != nil {
		util.Log(err)
		return nil
	}
	err = util.Unzip(dir, test.Data)
	if err != nil {
		util.Log(err)
		return nil
	}
	return &TestRunner{test.Project, test.Package, test.Names, test.Lang, dir}
}


//TestRunner is used to run tests on files compiled files.
type TestRunner struct {
	Project string
	Package string
	Names []string
	Lang    string
	Dir     string
}

//Execute sets up and runs tests on a compiled file. 
func (this *TestRunner)  Execute(f *submission.File, dir string) error {
	for _, name := range this.Names {
		target := tool.NewTarget(this.Project, name, this.Lang, this.Package, this.Dir)
		compiled, err := this.Compile(target, f, dir)
		if err != nil {
			return err
		}
		if !compiled {
			continue
		}
		err = this.Run(target, f, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

//Compile compiles a test for the current file. 
func (this *TestRunner) Compile(target *tool.TargetInfo, f *submission.File, dir string) (bool, error) {
	if _, ok := f.Results[target.Name+"_compile"]; ok {
		return true, nil
	}
	cp := dir+":"+this.Dir+":"+config.GetConfig(config.JUNIT_JAR)
	javac := tool.NewJavac(cp)
	res, err := javac.Run(f.Id, target)
	if err != nil{
		return false, err
	}
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}




//Run runs a test on the current file.
func (this *TestRunner) Run(target *tool.TargetInfo, f *submission.File, dir string) error {
	if _, ok := f.Results[target.Name+"_run"]; ok {
		return nil
	}
	junit := tool.NewJUnit(dir+":"+this.Dir, this.Dir)
	res, err := junit.Run(f.Id, target)
	if err != nil {
		return err
	}
	return AddResult(res)
}
