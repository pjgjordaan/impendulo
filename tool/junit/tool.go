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
		cp                   string
		dataLocation         string
		testInfo, runnerInfo *tool.TargetInfo
	}
)

//New creates a new  instance of the JUnit Tool.
//test is the JUnit Test to be run. dir is the location of the submission's tool directory.
func New(test *Test, dir string) (junit *Tool, err error) {
	//Load jar locations
	junitJar, err := config.JUNIT.Path()
	if err != nil {
		return
	}
	antJunit, err := config.ANT_JUNIT.Path()
	if err != nil {
		return
	}
	ant, err := config.ANT.Path()
	if err != nil {
		return
	}
	//Save the test files to the submission's tool directory.
	testInfo := tool.NewTarget(test.Name, test.Package, dir, tool.JAVA)
	err = util.SaveFile(testInfo.FilePath(), test.Test)
	if err != nil {
		return
	}
	if len(test.Data) != 0 {
		err = util.Unzip(testInfo.PackagePath(), test.Data)
		if err != nil {
			return
		}
	}
	dataLocation := filepath.Join(testInfo.PackagePath(), "data")
	//This is used to run the JUnit test using ant.
	runnerInfo := tool.NewTarget("TestRunner.java", "testing", testInfo.Dir, tool.JAVA)
	cp := testInfo.Dir + ":" + junitJar + ":" + antJunit + ":" + ant
	junit = &Tool{
		cp:           cp,
		dataLocation: dataLocation,
		testInfo:     testInfo,
		runnerInfo:   runnerInfo,
	}
	return
}

//Lang is Java
func (this *Tool) Lang() tool.Language {
	return tool.JAVA
}

//Name is JUnit
func (this *Tool) Name() string {
	return NAME
}

//Run runs a JUnit test on the provided Java source file. The source and test files are first
//compiled and we run the tests via a Java runner class which uses ant to generate XML output.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	java, err := config.JAVA.Path()
	if err != nil {
		return
	}
	cp := this.cp
	if cp != "" {
		cp += ":"
	}
	cp += ti.Dir
	comp, err := javac.New(cp)
	if err != nil {
		return
	}
	//First compile the files
	_, err = comp.Run(fileId, this.testInfo)
	if err != nil {
		return
	}
	_, err = comp.Run(fileId, this.runnerInfo)
	if err != nil {
		return
	}
	//Set the arguments
	outFile := filepath.Join(this.dataLocation, this.testInfo.Name+"_junit.xml")
	args := []string{java, "-cp", cp, this.runnerInfo.Executable(),
		this.testInfo.Executable(), this.dataLocation}
	defer os.Remove(outFile)
	//Run the tests and load the result
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, this.testInfo.Name, data)
		if err != nil && execRes.Err != nil {
			err = execRes.Err
		}
	} else if execRes.HasStdErr() {
		//The Java runner generated an error.
		err = fmt.Errorf("Could not run junit: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
