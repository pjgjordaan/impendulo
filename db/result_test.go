package db

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"html/template"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
)

func TestResult(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	file, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = Add(FILES, file)
	if err != nil {
		t.Error(err)
	}
	res := javac.NewResult(file.Id, fileData)
	err = AddResult(res)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": res.GetId()}
	dbRes, err := ToolResult(res.GetName(), matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(res, dbRes) {
		t.Error("Results not equivalent")
	}
}

func javacResult(fileId bson.ObjectId, gridFS bool) *javac.Result {
	id := bson.NewObjectId()
	return &javac.Result{
		Id:     id,
		FileId: fileId,
		Name:   javac.NAME,
		Data:   javacReport(id),
		GridFS: gridFS,
	}
}

func checkstyleResult(fileId bson.ObjectId, gridFS bool) *checkstyle.Result {
	id := bson.NewObjectId()
	return &checkstyle.Result{
		Id:     id,
		FileId: fileId,
		Name:   checkstyle.NAME,
		Data:   checkstyleReport(id),
		GridFS: gridFS,
	}
}

func findbugsResult(fileId bson.ObjectId, gridFS bool) *findbugs.Result {
	id := bson.NewObjectId()
	return &findbugs.Result{
		Id:     id,
		FileId: fileId,
		Name:   findbugs.NAME,
		Data:   findbugsReport(id),
		GridFS: gridFS,
	}
}

func report(name string, id bson.ObjectId) interface{} {
	switch name {
	case checkstyle.NAME:
		return checkstyleReport(id)
	case javac.NAME:
		return javacReport(id)
	case findbugs.NAME:
		return findbugsReport(id)
	default:
		return nil
	}
}

func findbugsReport(id bson.ObjectId) *findbugs.Report {
	report := new(findbugs.Report)
	report.Id = id
	report.Instances = []*findbugs.BugInstance{
		{
			Type:     "some bug",
			Priority: 1,
			Rank:     2,
		},
	}
	return report
}

func javacReport(id bson.ObjectId) *javac.Report {
	return &javac.Report{
		Id:    id,
		Type:  javac.ERRORS,
		Count: 4,
		Data:  []byte("some errors were found"),
	}
}

func checkstyleReport(id bson.ObjectId) *checkstyle.Report {
	return &checkstyle.Report{
		Id:      id,
		Version: "1.0",
		Errors:  0,
		Files: []*checkstyle.File{
			{
				Name: "Dummy.java",
				Errors: []*checkstyle.Error{
					{
						Line:     0,
						Column:   1,
						Severity: "Bad",
						Message:  template.HTML("<h2>Bad stuff</h2>"),
						Source:   "Some place.",
					},
				},
			},
		},
	}
}
