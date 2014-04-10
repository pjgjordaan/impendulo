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

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	//Tool is a tool.Tool used to run Tool tests on a Java source file.
	Tool struct {
		cp, name     string
		dataLocation string
		test, runner *tool.Target
	}
)

//New creates a new  instance of the JUnit Tool.
//test is the JUnit Test to be run. dir is the location of the submission's tool directory.
func New(test *Test, toolDir string) (*Tool, error) {
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
	td := filepath.Join(toolDir, test.Id.Hex())
	//Save the test files to the submission's tool directory.
	t := tool.NewTarget(test.Name, test.Package, td, tool.JAVA)
	if e = util.SaveFile(t.FilePath(), test.Test); e != nil {
		return nil, e
	}
	if len(test.Data) != 0 {
		if e = util.Unzip(t.PackagePath(), test.Data); e != nil {
			return nil, e
		}
	}
	return &Tool{
		cp:           toolDir + ":" + t.Dir + ":" + j + ":" + aj + ":" + a,
		dataLocation: filepath.Join(t.PackagePath(), "data"),
		test:         t,
		runner:       tool.NewTarget("TestRunner.java", "testing", toolDir, tool.JAVA),
	}, nil
}

//Lang is Java
func (t *Tool) Lang() tool.Language {
	return tool.JAVA
}

func (t *Tool) Name() string {
	return t.test.Name
}

//Run runs a JUnit test on the provided Java source file. The source and test files are first
//compiled and we run the tests via a Java runner class which uses ant to generate XML output.
func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (tool.ToolResult, error) {
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
	r, re := tool.RunCommand(a, nil)
	rf, e := os.Open(of)
	if e == nil {
		nr, e := NewResult(fileId, t.test.Name, util.ReadBytes(rf))
		if e != nil {
			if re != nil {
				e = re
			}
			return nil, e
		}
		return nr, nil
	} else if r.HasStdErr() {
		return nil, fmt.Errorf("could not run junit: %q.", string(r.StdErr))
	}
	return nil, re
}
