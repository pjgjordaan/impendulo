//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
	junitJar, err := config.JarFile(config.JUNIT)
	if err != nil {
		return
	}
	antJunit, err := config.JarFile(config.ANT_JUNIT)
	if err != nil {
		return
	}
	ant, err := config.JarFile(config.ANT)
	if err != nil {
		return
	}
	//Save the test files to the submission's tool directory.
	testInfo := tool.NewTarget(test.Name, tool.JAVA, test.Package, dir)
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
	runnerInfo := tool.NewTarget("TestRunner.java", "java", "testing", testInfo.Dir)
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
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name is JUnit
func (this *Tool) Name() string {
	return NAME
}

//Run runs a JUnit test on the provided Java source file. The source and test files are first
//compiled and we run the tests via a Java runner class which uses ant to generate XML output.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	java, err := config.Binary(config.JAVA)
	if err != nil {
		return
	}
	if this.cp != "" {
		this.cp += ":"
	}
	this.cp += ti.Dir
	comp, err := javac.New(this.cp)
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
	args := []string{java, "-cp", this.cp, this.runnerInfo.Executable(),
		this.testInfo.Executable(), this.dataLocation}
	defer os.Remove(outFile)
	//Run the tests and load the result
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, this.testInfo.Name, data)
	} else if execRes.HasStdErr() {
		//The Java runner generated an error.
		err = fmt.Errorf("Could not run junit: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
