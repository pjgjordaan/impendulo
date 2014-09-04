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

package file

import (
	"fmt"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	mk "github.com/godfried/impendulo/tool/make"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"path/filepath"
)

const (
	LOG_T = "processor/file/tools.go"
)

//Tools retrieves the Impendulo tool suite for a Worker's language.
//Each tool is already constructed.
func Tools(w *Worker) ([]tool.T, error) {
	switch tool.Language(w.project.Lang) {
	case tool.JAVA:
		return javaTools(w)
	case tool.C:
		return cTools(w), nil
	}
	//Only Java is supported so far...
	return nil, fmt.Errorf("no tools found for %s language", w.project.Lang)
}

func cTools(w *Worker) []tool.T {
	return []tool.T{}
}

//javaTools retrieves Impendulo's Java tool suite.
func javaTools(w *Worker) ([]tool.T, error) {
	a := make([]tool.T, 0, 10)
	//Only add tools if they were created successfully
	var t tool.T
	var e error
	t, e = checkstyle.New()
	if e != nil {
		return nil, e
	}
	a = append(a, t)
	t, e = findbugs.New()
	if e != nil {
		return nil, e
	}
	a = append(a, t)
	t, e = JPF(w)
	if e != nil {
		util.Log(e, LOG_T)
	} else {
		a = append(a, t)
	}
	t, e = PMD(w)
	if e != nil {
		return nil, e
	}
	a = append(a, t)
	ts, e := junitTools(w)
	if e != nil {
		return nil, e
	}
	return append(a, ts...), nil
}

//Compiler retrieves a compiler for a Processor's language.
func Compiler(w *Worker) (tool.Compiler, error) {
	l := tool.Language(w.project.Lang)
	switch l {
	case tool.JAVA:
		return javac.New("")
	case tool.C:
		m, e := db.Makefile(bson.M{db.PROJECTID: w.project.Id}, nil)
		if e != nil {
			return gcc.New()
		} else {
			return mk.New(m, w.toolDir)
		}
	}
	return nil, fmt.Errorf("no compiler found for %s language", l)
}

//JPF creates a new instance of the JPF tool.
func JPF(w *Worker) (tool.T, error) {
	//First we need the project's JPF configuration.
	c, e := db.JPFConfig(bson.M{db.PROJECTID: w.project.Id}, nil)
	if e != nil {
		return nil, e
	}
	return jpf.New(c, w.toolDir)
}

//PMD creates a new instance of the PMD tool.
func PMD(w *Worker) (tool.T, error) {
	//First we need the project's PMD rules.
	r, e := db.PMDRules(bson.M{db.PROJECTID: w.project.Id}, nil)
	if e != nil || r == nil || len(r.Rules) == 0 {
		r, e = pmd.DefaultRules(w.project.Id)
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

func junitTools(w *Worker) ([]tool.T, error) {
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: w.project.Id, db.TYPE: bson.M{db.NE: junit.USER}}, nil)
	if e != nil {
		return nil, e
	}
	d, e := config.JUNIT_TESTING.Path()
	if e != nil {
		return nil, e
	}
	if e = util.Copy(w.toolDir, d); e != nil {
		return nil, e
	}
	tools := make([]tool.T, 0, len(ts))
	for _, t := range ts {
		if t.Target == nil {
			continue
		}
		//Save the test files to the submission's tool directory.
		target := tool.NewTarget(t.Name, t.Package, filepath.Join(w.toolDir, t.Id.Hex()), tool.JAVA)
		if e = util.SaveFile(target.FilePath(), t.Test); e != nil {
			return nil, e
		}
		if len(t.Data) != 0 {
			if e = util.Unzip(target.PackagePath(), t.Data); e != nil {
				return nil, e
			}
		}
		ja, e := jacoco.New(w.rootDir, w.srcDir, target, t.Target, t.Id)
		if e != nil {
			return nil, e
		}
		tools = append(tools, ja)
		ju, e := junit.New(target, t.Target, w.toolDir, t.Id)
		if e != nil {
			return nil, e
		}
		tools = append(tools, ju)
	}
	return tools, nil
}
