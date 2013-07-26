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
//It creates and returns a new TestRunner for each available test.
func SetupTests(projectId bson.ObjectId, dir string) (ret []*TestRunner, err error) {
	proj, err := db.GetProject(bson.M{project.ID: projectId}, nil)
	if err != nil {
		return
	}
	tests, err := db.GetTests(bson.M{project.PROJECT_ID: projectId}, nil)
	if err != nil {
		return
	}
	ret = make([]*TestRunner, len(tests))
	for i, test := range tests {
		ret[i] = &TestRunner{tool.NewTarget(test.Name, proj.Lang, test.Package, dir)}
		err = util.SaveFile(ret[i].FilePath(), test.Test)
		if err != nil {
			return
		}
		if len(test.Data) == 0 {
			continue
		}
		err = util.Unzip(dir, test.Data)
		if err != nil {
			return
		}
	}
	return
}

//TestRunner is used to run tests on files compiled files.
type TestRunner struct {
	*tool.TargetInfo
}

//Run runs a test on the current file.
func (this *TestRunner) Run(f *project.File, srcDir string) error {
	ju := junit.NewJUnit(srcDir+":"+this.Dir, this.Dir)
	if _, ok := f.Results[this.Name]; ok {
		return nil
	}
	res, err := ju.Run(f.Id, this.TargetInfo)
	if err != nil {
		return err
	}
	return AddResult(res)
}
