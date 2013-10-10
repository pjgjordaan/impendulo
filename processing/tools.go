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
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

//Tools retrieves the Impendulo tool suite for a Processor's language.
//Each tool is already constructed.
func Tools(proc *Processor) (tools []tool.Tool, err error) {
	switch proc.project.Lang {
	case tool.JAVA:
		tools = javaTools(proc)
	default:
		//Only Java is supported so far...
		err = fmt.Errorf("No tools found for %s language.",
			proc.project.Lang)
	}
	return
}

//javaTools retrieves Impendulo's Java tool suite.
func javaTools(proc *Processor) []tool.Tool {
	tools := make([]tool.Tool, 0, 10)
	//Only add tools if they were created successfully
	fbTool, err := findbugs.New()
	if err == nil {
		tools = append(tools, fbTool)
	} else {
		util.Log(err)
	}
	csTool, err := checkstyle.New()
	if err == nil {
		tools = append(tools, csTool)
	} else {
		util.Log(err)
	}
	jpfTool, err := JPF(proc)
	if err == nil {
		tools = append(tools, jpfTool)
	} else {
		util.Log(err)
	}
	pmdTool, err := PMD(proc)
	if err == nil {
		tools = append(tools, pmdTool)
	} else {
		util.Log(err)
	}
	tests, err := JUnit(proc)
	if err == nil && len(tests) > 0 {
		tools = append(tools, tests...)
	} else {
		util.Log(err)
	}
	return tools
}

//Compiler retrieves a compiler for a Processor's language.
func Compiler(proc *Processor) (compiler tool.Tool, err error) {
	switch proc.project.Lang {
	case tool.JAVA:
		compiler, err = javac.New("")
	default:
		err = fmt.Errorf("No compiler found for %s language.",
			proc.project.Lang)
	}
	return
}

//JPF creates a new instance of the JPF tool.
func JPF(proc *Processor) (runnable tool.Tool, err error) {
	//First we need the project's JPF configuration.
	jpfFile, err := db.JPFConfig(
		bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	runnable, err = jpf.New(jpfFile, proc.toolDir)
	return
}

//PMD creates a new instance of the PMD tool.
func PMD(proc *Processor) (pmdTool tool.Tool, err error) {
	//First we need the project's PMD rules.
	rules, err := db.PMDRules(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil || rules == nil || len(rules.Rules) == 0 {
		rules, err = pmd.DefaultRules(proc.project.Id)
		if err != nil {
			return
		}
		err = db.AddPMDRules(rules)
	}
	pmdTool, err = pmd.New(rules)
	return
}

//JUnit creates a new JUnit tool instances for each available JUnit test for a given project.
func JUnit(proc *Processor) (ret []tool.Tool, err error) {
	//First we need the project's JUnit tests.
	tests, err := db.JUnitTests(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	testDir, err := config.JUNIT_TESTING.Path()
	if err != nil {
		return
	}
	//Now we copy our test runner to the proccessor's tool directory.
	err = util.Copy(proc.toolDir, testDir)
	if err != nil {
		return
	}
	ret = make([]tool.Tool, 0, len(tests))
	for _, test := range tests {
		unitTest, terr := junit.New(test, proc.toolDir)
		if terr != nil {
			util.Log(terr)
		} else {
			ret = append(ret, unitTest)
		}
	}
	return
}
