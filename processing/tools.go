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
	//These are the tools whose constructors don't return errors
	tools := []tool.Tool{
		findbugs.New(),
		checkstyle.New(),
	}
	//Only add JPF if it was created successfully
	jpfTool, err := JPF(proc)
	if err == nil {
		tools = append(tools, jpfTool)
	} else {
		util.Log(err)
	}
	//ditto PMD and JUnit
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
		compiler = javac.New("")
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
func PMD(proc *Processor) (runnable tool.Tool, err error) {
	//First we need the project's PMD rules.
	rules, err := db.PMDRules(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		rules, err = pmd.DefaultRules(proc.project.Id)
		if err != nil {
			return
		}
		err = db.AddPMDRules(rules)
	}
	runnable = pmd.New(rules)
	return
}

//JUnit creates a new JUnit tool instances for each available JUnit test for a given project.
func JUnit(proc *Processor) (ret []tool.Tool, err error) {
	//First we need the project's JUnit tests.
	tests, err := db.JUnitTests(bson.M{project.PROJECT_ID: proc.project.Id}, nil)
	if err != nil {
		return
	}
	//Now we copy our test runner to the proccessor's tool directory.
	err = util.Copy(proc.toolDir, config.Config(config.TESTING_DIR))
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
