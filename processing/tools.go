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

package processing

import (
	"fmt"
	"path/filepath"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/junit_user"
	mk "github.com/godfried/impendulo/tool/make"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

const (
	LOG_TOOLS = "processing/tools.go"
)

func TestTools(p *TestProcessor, tf *project.File) ([]tool.Tool, error) {
	switch tool.Language(p.project.Lang) {
	case tool.JAVA:
		return javaTestTools(p, tf), nil
	case tool.C:
		return cTestTools(p, tf), nil
	}
	//Only Java is supported so far...
	return nil, fmt.Errorf("no tools found for %s language", p.project.Lang)
}

func cTestTools(p *TestProcessor, tf *project.File) []tool.Tool {
	return []tool.Tool{}
}

func javaTestTools(p *TestProcessor, tf *project.File) []tool.Tool {
	a := make([]tool.Tool, 0, 10)
	test := &junit.Test{
		Id:      tf.Id,
		Name:    tf.Name,
		Package: tf.Package,
		Test:    tf.Data,
	}
	var t tool.Tool
	var e error
	t, e = junit.New(test, p.toolDir)
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else {
		a = append(a, t)
	}
	t, e = Jacoco(p, test)
	if e != nil {
		util.Log(e, LOG_TOOLS)
		return a
	}
	return append(a, t)
}

func Jacoco(p *TestProcessor, test *junit.Test) (tool.Tool, error) {
	t := tool.NewTarget(test.Name, test.Package, filepath.Join(p.toolDir, test.Id.Hex()), tool.JAVA)
	return jacoco.New(p.rootDir, p.srcDir, t)
}

//Tools retrieves the Impendulo tool suite for a Processor's language.
//Each tool is already constructed.
func Tools(p *Processor) ([]tool.Tool, error) {
	switch tool.Language(p.project.Lang) {
	case tool.JAVA:
		return javaTools(p), nil
	case tool.C:
		return cTools(p), nil
	}
	//Only Java is supported so far...
	return nil, fmt.Errorf("no tools found for %s language", p.project.Lang)
}

func cTools(p *Processor) []tool.Tool {
	return []tool.Tool{}
}

//javaTools retrieves Impendulo's Java tool suite.
func javaTools(p *Processor) []tool.Tool {
	a := make([]tool.Tool, 0, 10)
	//Only add tools if they were created successfully
	var t tool.Tool
	var e error
	t, e = checkstyle.New()
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else {
		a = append(a, t)
	}
	t, e = findbugs.New()
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else {
		a = append(a, t)
	}
	t, e = JPF(p)
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else {
		a = append(a, t)
	}
	t, e = PMD(p)
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else {
		a = append(a, t)
	}
	tests, e := JUnit(p)
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else if len(tests) > 0 {
		a = append(a, tests...)
	}
	tests, e = UserJUnit(p)
	if e != nil {
		util.Log(e, LOG_TOOLS)
	} else if len(tests) > 0 {
		a = append(a, tests...)
	}
	return a
}

//Compiler retrieves a compiler for a Processor's language.
func Compiler(p *Processor) (tool.Tool, error) {
	l := tool.Language(p.project.Lang)
	switch l {
	case tool.JAVA:
		return javac.New("")
	case tool.C:
		m, e := db.Makefile(bson.M{db.PROJECTID: p.project.Id}, nil)
		if e != nil {
			return gcc.New()
		} else {
			return mk.New(m, p.toolDir)
		}
	}
	return nil, fmt.Errorf("no compiler found for %s language", l)
}

//JPF creates a new instance of the JPF tool.
func JPF(p *Processor) (tool.Tool, error) {
	//First we need the project's JPF configuration.
	c, e := db.JPFConfig(bson.M{db.PROJECTID: p.project.Id}, nil)
	if e != nil {
		return nil, e
	}
	return jpf.New(c, p.toolDir)
}

//PMD creates a new instance of the PMD tool.
func PMD(p *Processor) (tool.Tool, error) {
	//First we need the project's PMD rules.
	r, e := db.PMDRules(bson.M{db.PROJECTID: p.project.Id}, nil)
	if e != nil || r == nil || len(r.Rules) == 0 {
		r, e = pmd.DefaultRules(p.project.Id)
		if e != nil {
			return nil, e
		}
		e = db.AddPMDRules(r)
		if e != nil {
			return nil, e
		}
	}
	return pmd.New(r)
}

//JUnit creates a new JUnit tool instances for each available JUnit test for a given project.
func JUnit(p *Processor) ([]tool.Tool, error) {
	//First we need the project's JUnit tests.
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: p.project.Id, db.TYPE: bson.M{db.NE: junit.USER}}, nil)
	if e != nil {
		return nil, e
	}
	d, e := config.JUNIT_TESTING.Path()
	if e != nil {
		return nil, e
	}
	//Now we copy our test runner to the proccessor's tool directory.
	if e = util.Copy(p.toolDir, d); e != nil {
		return nil, e
	}
	js := make([]tool.Tool, 0, len(ts))
	for _, t := range ts {
		j, e := junit.New(t, p.toolDir)
		if e != nil {
			util.Log(e, LOG_TOOLS)
		} else {
			js = append(js, j)
		}
	}
	return js, nil
}

func UserJUnit(p *Processor) ([]tool.Tool, error) {
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: p.project.Id, db.TYPE: junit.USER}, nil)
	if e != nil {
		return nil, e
	}
	d, e := config.JUNIT_TESTING.Path()
	if e != nil {
		return nil, e
	}
	if e = util.Copy(p.toolDir, d); e != nil {
		return nil, e
	}
	js := make([]tool.Tool, 0, len(ts))
	for _, t := range ts {
		js = append(js, junit_user.New(t, p.toolDir))
	}
	return js, nil
}
