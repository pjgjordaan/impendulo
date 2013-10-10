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

//Package pmd is the PMD static analysis tool's implementation of an Impendulo tool.
//For more information see http://pmd.sourceforge.net/.
package pmd

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"strings"
)

type (
	//Tool is an implementation of tool.Tool which allows us to run
	//PMD on Java classes.
	Tool struct {
		cmd   string
		rules string
	}
)

//New creates a new instance of a PMD Tool.
//Errors returned will be due to loading either the default
//PMD rules or the PMD execution script.
func New(rules *Rules) (tool *Tool, err error) {
	if rules == nil {
		rules, err = DefaultRules(bson.NewObjectId())
		if err != nil {
			return
		}
	}
	tool = &Tool{
		rules: strings.Join(rules.RuleArray(), ","),
	}
	tool.cmd, err = config.PMD.Path()
	return
}

//Lang is Java
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name is PMD
func (this *Tool) Name() string {
	return NAME
}

//Run runs PMD on a provided Java source file. PMD writes its output to an XML file which we then read
//and use to create a PMD Result.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "pmd.xml")
	args := []string{this.cmd, "pmd", "-f", "xml", "-stress",
		"-shortnames", "-R", this.rules, "-r", outFile, "-d", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run pmd: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
