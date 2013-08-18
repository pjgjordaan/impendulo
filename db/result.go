package db

import (
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
	"strings"
	"sort"
)

//GetCheckstyleResult retrieves a CheckstyleResult matching
//the given interface from the active database.
func GetCheckstyleResult(matcher, selector bson.M) (ret *checkstyle.CheckstyleResult, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = checkstyle.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//GetPMDResult retrieves a PMDResult matching
//the given interface from the active database.
func GetPMDResult(matcher, selector bson.M) (ret *pmd.PMDResult, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = pmd.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//GetFindbugsResult retrieves a FindbugsResult matching
//the given interface from the active database.
func GetFindbugsResult(matcher, selector bson.M) (ret *findbugs.FindbugsResult, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = findbugs.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//GetJPFResult retrieves a JPFResult matching
//the given interface from the active database.
func GetJPFResult(matcher, selector bson.M) (ret *jpf.JPFResult, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = jpf.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//GetJUnitResult retrieves a JUnitResult matching
//the given interface from the active database.
func GetJUnitResult(matcher, selector bson.M) (ret *junit.JUnitResult, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//GetJavacResult retrieves a JavacResult matching
//the given interface from the active database.
func GetJavacResult(matcher, selector bson.M) (ret *javac.JavacResult, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = javac.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//GetToolResult retrieves a tool.ToolResult matching
//the given interface and name from the active database.
func GetToolResult(name string, matcher, selector bson.M) (ret tool.ToolResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = GetJavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = GetJPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = GetFindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = GetPMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = GetCheckstyleResult(matcher, selector)
	default:
		ret, err = GetJUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//GetDisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func GetDisplayResult(name string, matcher, selector bson.M) (ret tool.DisplayResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = GetJavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = GetJPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = GetFindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = GetPMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = GetCheckstyleResult(matcher, selector)
	default:
		ret, err = GetJUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//GetDisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func GetGraphResult(name string, matcher, selector bson.M) (ret tool.GraphResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = GetJavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = GetJPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = GetFindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = GetPMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = GetCheckstyleResult(matcher, selector)
	default:
		ret, err = GetJUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(res tool.ToolResult) (err error) {
	if res == nil{
		err = fmt.Errorf("Result is nil. In db/result.go.")
		return
	}
	matcher := bson.M{project.ID: res.GetFileId()}
	change := bson.M{SET: bson.M{project.RESULTS + "." + res.GetName(): res.GetId()}}
	err = Update(FILES, matcher, change)
	if err != nil {
		return
	}
	session, err := getSession()
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

//AddTimeoutResult adds a new TimeoutResult to the active database.
func AddTimeoutResult(fileId bson.ObjectId, name string) (err error) {
	matcher := bson.M{project.ID: fileId}
	change := bson.M{SET: bson.M{project.RESULTS + "." + name: tool.TIMEOUT}}
	err = Update(FILES, matcher, change)
	return
}

//AddNoResult adds a new NoResult to the active database.
func AddNoResult(fileId bson.ObjectId, name string) (err error) {
	matcher := bson.M{project.ID: fileId}
	change := bson.M{SET: bson.M{project.RESULTS + "." + name: tool.NORESULT}}
	err = Update(FILES, matcher, change)
	return
}

//GetAllResults retrieves all tool.DisplayResults matching
//the given file Id from the active database.
func GetGraphResults(fileId bson.ObjectId) (ret []tool.GraphResult, err error) {
	file, err := GetFile(bson.M{project.ID: fileId}, bson.M{project.RESULTS: 1})
	if err != nil{
		return
	}
	ret = make([]tool.GraphResult, 0)
	for name, id := range file.Results{
		if _, ok := id.(bson.ObjectId); !ok{
			continue
		}
		res, err := GetGraphResult(name, bson.M{project.ID: id}, nil) 
		if err != nil{
			err = nil
			continue
		}
		ret = append(ret, res)
	}
	return
}

//GetResultNames retrieves all result names for a given project.
func GetResultNames(projectId bson.ObjectId, nonTool bool) (ret []string, err error) {
	tests, err := GetTests(bson.M{project.PROJECT_ID: projectId},
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
	if nonTool{
		ret = append(ret, tool.CODE, tool.SUMMARY)
	}
	sort.Strings(ret)
	return
}

//RemoveResultById removes a result matching
//the given id from the active database.
func RemoveResultById(id interface{}) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"result", err, id}
	}
	return
}
