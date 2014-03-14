//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package db

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	junit_user "github.com/godfried/impendulo/tool/junit_user/result"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo/bson"
	"sort"
	"strings"
)

type (
	TypeHolder struct {
		Type string "type"
	}
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
	matcher[NAME] = checkstyle.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
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
	matcher[NAME] = pmd.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
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
	matcher[NAME] = findbugs.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
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
	matcher[NAME] = jpf.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
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
	matcher[TYPE] = junit.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func JacocoResult(matcher, selector bson.M) (ret *jacoco.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[TYPE] = jacoco.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func JUnitUserResult(matcher, selector bson.M) (ret *junit_user.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[TYPE] = junit_user.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
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
	matcher[NAME] = javac.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func GCCResult(matcher, selector bson.M) (ret *gcc.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = gcc.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func resultType(matcher bson.M) (string, error) {
	session, err := Session()
	if err != nil {
		return "", err
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	var holder *TypeHolder
	err = c.Find(matcher).One(&holder)
	if err != nil {
		return "", &DBGetError{"result type", err, matcher}
	}
	return holder.Type, nil
}

//ToolResult retrieves a tool.ToolResult matching
//the given interface and name from the active database.
func ToolResult(matcher, selector bson.M) (ret tool.ToolResult, err error) {
	tipe, err := resultType(matcher)
	if err != nil {
		return
	}
	switch tipe {
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
	case gcc.NAME:
		ret, err = GCCResult(matcher, selector)
	case jacoco.NAME:
		ret, err = JacocoResult(matcher, selector)
	case junit.NAME:
		ret, err = JUnitResult(matcher, selector)
	case junit_user.NAME:
		ret, err = JUnitUserResult(matcher, selector)
	default:
		err = fmt.Errorf("Unsupported result type %s.", tipe)
	}
	return
}

//DisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func DisplayResult(matcher, selector bson.M) (ret tool.DisplayResult, err error) {
	tipe, err := resultType(matcher)
	if err != nil {
		return
	}
	switch tipe {
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
	case gcc.NAME:
		ret, err = GCCResult(matcher, selector)
	case jacoco.NAME:
		ret, err = JacocoResult(matcher, selector)
	case junit.NAME:
		ret, err = JUnitResult(matcher, selector)
	case junit_user.NAME:
		ret, err = JUnitUserResult(matcher, selector)
	default:
		err = fmt.Errorf("Unsupported result type %s.", tipe)
	}
	return
}

func ChartResult(matcher, selector bson.M) (ret tool.ChartResult, err error) {
	tipe, err := resultType(matcher)
	if err != nil {
		return
	}
	switch tipe {
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
	case gcc.NAME:
		ret, err = GCCResult(matcher, selector)
	case junit.NAME:
		ret, err = JUnitResult(matcher, selector)
	case jacoco.NAME:
		ret, err = JacocoResult(matcher, selector)
	case junit_user.NAME:
		ret, err = JUnitUserResult(matcher, selector)
	default:
		err = fmt.Errorf("Unsupported result type %s.", tipe)
	}
	return
}

func ResultName(matcher bson.M) (name string, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	var names []string
	err = c.Find(matcher).
		Select(bson.M{NAME: 1}).
		Distinct(NAME, &names)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if len(names) == 0 {
		err = fmt.Errorf("Could not retrieve result name.")
	} else {
		name = names[0]
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(res tool.ToolResult, name string) (err error) {
	if res == nil {
		err = fmt.Errorf("Result is nil. In db/result.go.")
		return
	}
	err = AddFileResult(res.GetFileId(), name, res.GetId())
	if err != nil {
		return
	}
	if res.OnGridFS() {
		err = AddGridFile(res.GetId(), res.GetReport())
		if err != nil {
			return
		}
		res.SetReport(nil)
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

//AddFileResult adds or updates a result in a file's results.
func AddFileResult(fileId bson.ObjectId, name string, value interface{}) error {
	matcher := bson.M{ID: fileId}
	change := bson.M{SET: bson.M{RESULTS + "." + name: value}}
	return Update(FILES, matcher, change)
}

//ChartResults retrieves all tool.DisplayResults matching
//the given file Id from the active database.
func ChartResults(fileId bson.ObjectId) (ret []tool.ChartResult, err error) {
	file, err := File(bson.M{ID: fileId}, bson.M{DATA: 0})
	if err != nil {
		return
	}
	ret = make([]tool.ChartResult, 0, len(file.Results))
	for _, id := range file.Results {
		if _, ok := id.(bson.ObjectId); !ok {
			continue
		}
		res, err := ChartResult(bson.M{ID: id}, nil)
		if err != nil {
			err = nil
			continue
		}
		ret = append(ret, res)
	}
	return
}

//AllResultNames retrieves all result names for a given project.
func AllResultNames(projectId bson.ObjectId) (ret []string, err error) {
	ret, err = ToolResultNames(projectId)
	if err != nil {
		return
	}
	ret = append(ret, tool.CODE, diff.NAME, tool.SUMMARY)
	sort.Strings(ret)
	return
}

//ChartResultNames retrieves all result names for a given project.
func ChartResultNames(projectId bson.ObjectId) (ret []string, err error) {
	ret, err = ToolResultNames(projectId)
	if err != nil {
		return
	}
	ret = append(ret, tool.SUMMARY)
	sort.Strings(ret)
	return
}

func ToolResultNames(projectId bson.ObjectId) (ret []string, err error) {
	p, err := Project(bson.M{ID: projectId}, bson.M{LANG: 1})
	if err != nil {
		return
	}
	switch tool.Language(p.Lang) {
	case tool.JAVA:
		ret, err = JavaResultNames(projectId)
	case tool.C:
		ret, err = CResultNames(projectId)
	}
	return
}

func JavaResultNames(projectId bson.ObjectId) (ret []string, err error) {
	tests, err := JUnitTests(bson.M{PROJECTID: projectId}, bson.M{NAME: 1})
	if err != nil {
		return
	}
	ret = make([]string, len(tests))
	for i, test := range tests {
		ret[i] = strings.Split(test.Name, ".")[0]
	}
	ret = append(ret, checkstyle.NAME, findbugs.NAME, javac.NAME)
	_, jerr := JPFConfig(bson.M{PROJECTID: projectId}, bson.M{ID: 1})
	if jerr == nil {
		ret = append(ret, jpf.NAME)
	}
	_, perr := PMDRules(bson.M{PROJECTID: projectId}, bson.M{ID: 1})
	if perr == nil {
		ret = append(ret, pmd.NAME)
	}
	sort.Strings(ret)
	return
}

func CResultNames(projectId bson.ObjectId) (ret []string, err error) {
	ret = []string{gcc.NAME}
	return
}
