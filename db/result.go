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
)

//GetCheckstyleResult retrieves a CheckstyleResult matching
//the given interface from the active database.
func GetCheckstyleResult(matcher, selector interface{}) (ret *checkstyle.CheckstyleResult, err error) {
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

//GetPMDResult retrieves a PMDResult matching
//the given interface from the active database.
func GetPMDResult(matcher, selector interface{}) (ret *pmd.PMDResult, err error) {
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

//GetFindbugsResult retrieves a FindbugsResult matching
//the given interface from the active database.
func GetFindbugsResult(matcher, selector interface{}) (ret *findbugs.FindbugsResult, err error) {
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

//GetJPFResult retrieves a JPFResult matching
//the given interface from the active database.
func GetJPFResult(matcher, selector interface{}) (ret *jpf.JPFResult, err error) {
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

//GetJUnitResult retrieves a JUnitResult matching
//the given interface from the active database.
func GetJUnitResult(matcher, selector interface{}) (ret *junit.JUnitResult, err error) {
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
func GetJavacResult(matcher, selector interface{}) (ret *javac.JavacResult, err error) {
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

//GetToolResult retrieves a tool.ToolResult matching
//the given interface and name from the active database.
func GetToolResult(name string, matcher, selector interface{}) (ret tool.ToolResult, err error) {
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
func GetDisplayResult(name string, matcher, selector interface{}) (ret tool.DisplayResult, err error) {
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

//GetResultNames retrieves all result names for a given project.
func GetResultNames(projectId bson.ObjectId) (ret []string, err error) {
	tests, err := GetTests(bson.M{project.PROJECT_ID: projectId},
		bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	ret = make([]string, len(tests))
	for i, test := range tests {
		ret[i] = strings.Split(test.Name, ".")[0]
	}
	ret = append(ret, []string{tool.CODE, checkstyle.NAME, findbugs.NAME,
		javac.NAME, jpf.NAME, pmd.NAME, tool.SUMMARY}...)
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
