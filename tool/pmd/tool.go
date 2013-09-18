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

//New
func New(rules *Rules) *Tool {
	if rules == nil {
		rules, _ = DefaultRules(bson.NewObjectId())
	}
	return &Tool{
		cmd:   config.Config(config.PMD),
		rules: strings.Join(rules.RuleArray(), ","),
	}
}

//Lang
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name
func (this *Tool) Name() string {
	return NAME
}

//Run runs PMD on a provided Java source file. PMD writes its output to an XML file which we then read
//and use to create a PMD Result.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "pmd.xml")
	args := []string{this.cmd, config.PMD, "-f", "xml", "-stress",
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
