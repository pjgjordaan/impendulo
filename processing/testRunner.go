package processing

import (
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"strings"
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
	return &TestRunner{test.Project, test.Names, test.Lang, dir}
}


//TestRunner is used to run tests on files compiled files.
type TestRunner struct {
	Project string
	Names []string
	Lang    string
	Dir     string
}

//Execute sets up and runs tests on a compiled file. 
func (this *TestRunner)  Execute(f *submission.File, target *tool.TargetInfo) error {
	for _, name := range this.Names {
		compiled, err := this.Compile(name, f, target)
		if err != nil {
			return err
		}
		if !compiled {
			continue
		}
		err = this.Run(name, f, target)
		if err != nil {
			return err
		}
	}
	return nil
}

const(
	JUNIT_EXEC = "org.junit.runner.JUnitCore"
	JUNIT_JAR = "/usr/share/java/junit4.jar"
)

//Compile compiles a test for the current file. 
func (this *TestRunner) Compile(testName string, f *submission.File, target *tool.TargetInfo) (bool, error) {
	if _, ok := f.Results[testName+"_compile"]; ok {
		return true, nil
	}
	cp := this.Dir+":"+JUNIT_JAR
	stderr, stdout, ok, err := tool.RunCommand("javac", "-cp", cp, "-implicit:class", filepath.Join(this.Dir,"testing",testName))
	if !ok{
		return false, err
	}
	res := tool.NewResult(f.Id, f.Id, testName+"_compile", "warnings", "errors", stdout, stderr, err)
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}

//Run runs a test on the current file.
func (this *TestRunner) Run(testName string, f *submission.File, target *tool.TargetInfo) error {
	if _, ok := f.Results[testName+"_run"]; ok {
		return nil
	}
	cp := this.Dir+":"+JUNIT_JAR
	env := "-Ddata.location="+this.Dir
	exec := strings.Split(testName, ".")[0]
	stderr, stdout, ok, err := tool.RunCommand("java", "-cp", cp, env, JUNIT_EXEC, "testing."+exec)
	if !ok {
		return err
	}
	res := tool.NewResult(f.Id, f.Id, testName+"_run", "warnings", "errors", stdout, stderr, err)
	return AddResult(res)
}
