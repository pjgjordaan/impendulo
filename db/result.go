package db

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"strings"
	"fmt"
	"labix.org/v2/mgo/bson"
)

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

func GetResult(name string, matcher, selector interface{}) (ret tool.Result, err error) {
	switch name{
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
		if err != nil{
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(r tool.Result) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err = col.Insert(r)
	if err != nil {
		err = &DBAddError{r.String(), err}
	}
	return
}

func GetResultNames(projectId bson.ObjectId) (ret []string, err error) {
	tests, err := GetTests(bson.M{project.PROJECT_ID: projectId}, bson.M{project.NAME:1})
	if err != nil{
		return
	}
	ret = make([]string, len(tests))
	for i, test := range tests{
		ret[i] = strings.Split(test.Name, ".")[0]
	}
	ret = append(ret, []string{tool.CODE, checkstyle.NAME, findbugs.NAME, javac.NAME, jpf.NAME, pmd.NAME}...) 
	return 
}
