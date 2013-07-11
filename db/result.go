package db

import (
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/findbugs"
)

func GetFindbugsResult(matcher, selector interface{}) (ret *findbugs.FindbugsResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}


func GetJPFResult(matcher, selector interface{}) (ret *jpf.JPFResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetJUnitResult(matcher, selector interface{}) (ret *junit.JUnitResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetJavacResult(matcher, selector interface{}) (ret *javac.JavacResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}


//AddResult adds a new result to the active database.
func AddResult(r tool.Result)(err error){
	session := getSession()
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err = col.Insert(r)
	if err != nil {
		err = &DBAddError{r.String(), err}
	}
	return
}
