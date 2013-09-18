package db

import (
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
	"sort"
	"strings"
)

//CheckstyleResult retrieves a Result matching
//the given interface from the active database.
func CheckstyleResult(matcher, selector bson.M) (ret *checkstyle.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = checkstyle.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Data)
	}
	return
}

//PMDResult retrieves a Result matching
//the given interface from the active database.
func PMDResult(matcher, selector bson.M) (ret *pmd.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = pmd.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Data)
	}
	return
}

//FindbugsResult retrieves a Result matching
//the given interface from the active database.
func FindbugsResult(matcher, selector bson.M) (ret *findbugs.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = findbugs.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Data)
	}
	return
}

//JPFResult retrieves a Result matching
//the given interface from the active database.
func JPFResult(matcher, selector bson.M) (ret *jpf.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = jpf.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Data)
	}
	return
}

//JUnitResult retrieves aResult matching
//the given interface from the active database.
func JUnitResult(matcher, selector bson.M) (ret *junit.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Data)
	}
	return
}

//JavacResult retrieves a JavacResult matching
//the given interface from the active database.
func JavacResult(matcher, selector bson.M) (ret *javac.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = javac.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Data)
		if err != nil {
			return
		}
	}
	return
}

//ToolResult retrieves a tool.ToolResult matching
//the given interface and name from the active database.
func ToolResult(name string, matcher, selector bson.M) (ret tool.ToolResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = JavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = JPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = FindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = PMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = CheckstyleResult(matcher, selector)
	default:
		ret, err = JUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//DisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func DisplayResult(name string, matcher, selector bson.M) (ret tool.DisplayResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = JavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = JPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = FindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = PMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = CheckstyleResult(matcher, selector)
	default:
		ret, err = JUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//DisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func GraphResult(name string, matcher, selector bson.M) (ret tool.GraphResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = JavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = JPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = FindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = PMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = CheckstyleResult(matcher, selector)
	default:
		ret, err = JUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(res tool.ToolResult) (err error) {
	if res == nil {
		err = fmt.Errorf("Result is nil. In db/result.go.")
		return
	}
	err = AddFileResult(res.GetFileId(), res.GetName(), res.GetId())
	if err != nil {
		return
	}
	if res.OnGridFS() {
		err = AddGridFile(res.GetId(), res.GetData())
		if err != nil {
			return
		}
		res.SetData(nil)
	}
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err = col.Insert(res)
	if err != nil {
		err = &DBAddError{res.GetName(), err}
	}
	return
}

func AddFileResult(fileId bson.ObjectId, name string, value interface{}) error {
	matcher := bson.M{project.ID: fileId}
	change := bson.M{SET: bson.M{project.RESULTS + "." + name: value}}
	return Update(FILES, matcher, change)
}

//GraphResults retrieves all tool.DisplayResults matching
//the given file Id from the active database.
func GraphResults(fileId bson.ObjectId) (ret []tool.GraphResult, err error) {
	file, err := File(bson.M{project.ID: fileId}, bson.M{project.RESULTS: 1})
	if err != nil {
		return
	}
	ret = make([]tool.GraphResult, 0, len(file.Results))
	for name, id := range file.Results {
		if _, ok := id.(bson.ObjectId); !ok {
			continue
		}
		res, err := GraphResult(name, bson.M{project.ID: id}, nil)
		if err != nil {
			err = nil
			continue
		}
		ret = append(ret, res)
	}
	return
}

//ResultNames retrieves all result names for a given project.
func ResultNames(projectId bson.ObjectId, nonTool bool) (ret []string, err error) {
	tests, err := JUnitTests(bson.M{project.PROJECT_ID: projectId},
		bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	ret = make([]string, len(tests))
	for i, test := range tests {
		ret[i] = strings.Split(test.Name, ".")[0]
	}
	ret = append(ret, checkstyle.NAME, findbugs.NAME,
		javac.NAME, jpf.NAME, pmd.NAME)
	if nonTool {
		ret = append(ret, tool.CODE, diff.NAME, tool.SUMMARY)
	}
	sort.Strings(ret)
	return
}
