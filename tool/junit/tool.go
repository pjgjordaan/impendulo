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

//Package JUnit is the JUnit Java testing framework's implementation of an Impendulo tool.
//See http://junit.org/ for more information.
package junit

import (
	"fmt"
	"time"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	//Tool is a tool.T used to run Tool tests on a Java source file.
	Tool struct {
		cp, name             string
		dataLocation         string
		test, target, runner *tool.Target
		testId               bson.ObjectId
	}
)

//New creates a new  instance of the JUnit Tool.
//test is the JUnit Test to be run. dir is the location of the submission's tool directory.
func New(test, target *tool.Target, toolDir string, testId bson.ObjectId) (*Tool, error) {
	//Load jar locations
	j, e := config.JUNIT.Path()
	if e != nil {
		return nil, e
	}
	aj, e := config.ANT_JUNIT.Path()
	if e != nil {
		return nil, e
	}
	a, e := config.ANT.Path()
	if e != nil {
		return nil, e
	}
	return &Tool{
		cp:           toolDir + ":" + test.Dir + ":" + j + ":" + aj + ":" + a,
		dataLocation: filepath.Join(test.PackagePath(), "data"),
		test:         test,
		runner:       tool.NewTarget("TestRunner.java", "testing", toolDir, project.JAVA),
		target:       target,
		testId:       testId,
	}, nil
}

//Lang is Java
func (t *Tool) Lang() project.Language {
	return project.JAVA
}

func (t *Tool) Name() string {
	return NAME + ":" + t.test.Name
}

//Run runs a JUnit test on the provided Java source file. The source and test files are first
//compiled and we run the tests via a Java runner class which uses ant to generate XML output.
func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (result.Tooler, error) {
	if t.target.Executable() != target.Executable() {
		return nil, fmt.Errorf("file executable %s does not match expected executable %s", target.Executable(), t.target.Executable())
	}
	jp, e := config.JAVA.Path()
	if e != nil {
		return nil, e
	}
	cp := t.cp
	if cp != "" {
		cp += ":"
	}
	cp += target.Dir
	c, e := javac.New(cp)
	if e != nil {
		return nil, e
	}
	//First compile the files
	if _, e = c.Run(fileId, t.test); e != nil {
		return nil, e
	}
	if _, e = c.Run(fileId, t.runner); e != nil {
		return nil, e
	}
	//Set the arguments
	on := t.test.Name + "_junit"
	od := target.PackagePath()
	of := filepath.Join(od, t.test.Name+"_junit.xml")
	a := []string{jp, "-cp", cp, t.runner.Executable(), t.test.Executable(), t.dataLocation, on, od}
	defer os.Remove(of)
	//Run the tests and load the result
	r, re := tool.RunCommand(a, nil, 30*time.Second)
	rf, oe := os.Open(of)
	if oe != nil {
		if re != nil {
			return nil, re
		}
		return nil, fmt.Errorf("could not run junit: %q.", string(r.StdErr))
	}
	defer rf.Close()
	nr, e := NewResult(fileId, t.testId, t.test.Name, util.ReadBytes(rf))
	if e != nil {
		if re != nil {
			e = re
		}
		return nil, e
	}
	return nr, nil
}
