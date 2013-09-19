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

//Package findbugs is the Findbugs static analysis tool's implementation of an Imendulo tool.
//See http://findbugs.sourceforge.net/ for more information.
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
	//Findbugs is a tool.Tool used to run Findbugs on Java classes.
	FindBugs struct {
		cmd string
	}
)

//New creates a new instance of the Findbugs tool.
//If an error is returned, it will be due Findbugs not being configured correctly.
func New() (tool *FindBugs, err error) {
	tool = new(FindBugs)
	tool.cmd, err = config.JarFile(config.FINDBUGS)
	return
}

//Lang is Java.
func (this *FindBugs) Lang() string {
	return tool.JAVA
}

//Name is Findbugs.
func (this *FindBugs) Name() string {
	return NAME
}

//Run executes Findbugs on the provided source file.
//Findbugs is run with the following flags: -effort:max, -experimental, -relaxed.
//The result is written to an XML file which is then read and used to create a
//Findbugs Result.
func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	//Setup arguments
	java, err := config.Binary(config.JAVA)
	if err != nil {
		return
	}
	outFile := filepath.Join(ti.Dir, "findbugs.xml")
	args := []string{java, "-jar", this.cmd, "-textui", "-effort:max", "-experimental",
		"-xml:withMessages", "-relaxed", "-output", outFile, ti.PackagePath()}
	defer os.Remove(outFile)
	//Run Findbugs and load result.
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Success
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		//Findbugs generated an error
		err = fmt.Errorf("Could not run findbugs: %q.",
			string(execRes.StdErr))
	} else {
		//Normal error
		err = execRes.Err
	}
	return
}
