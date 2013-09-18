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

package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

type (
	//Findbugs is a tool.Tool used to run the Findbugs static analysis tool on
	//Java classes.
	FindBugs struct {
		cmd string
	}
)

//New Creates a new instance of the Findbugs tool.
func New() *FindBugs {
	return &FindBugs{config.Config(config.FINDBUGS)}
}

//Lang
func (this *FindBugs) Lang() string {
	return tool.JAVA
}

//Name
func (this *FindBugs) Name() string {
	return NAME
}

//Run
func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "findbugs.xml")
	args := []string{config.Config(config.JAVA), "-jar", this.cmd, "-textui", "-effort:max",
		"-experimental", "-xml:withMessages", "-relaxed", "-output", outFile, ti.PackagePath()}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run findbugs: %q.",
			string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
