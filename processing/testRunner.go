package processing

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
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
		err = util.SaveFile(ret[i].Info.FilePath(), test.Test)
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

//Run runs a test on the current file.
func (this *TestRunner) Run(f *project.File, srcDir string) error {
	ju := junit.NewJUnit(srcDir+":"+this.Info.Dir, this.Info.Dir)
	if _, ok := f.Results[this.Info.Name+"_"+ju.GetName()]; ok {
		return nil
	}
	res, err := ju.Run(f.Id, this.Info)
	util.Log("Test run result", res)
	if err != nil {
		return err
	}
	return AddResult(res)
}
