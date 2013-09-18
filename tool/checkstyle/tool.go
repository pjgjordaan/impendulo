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

package checkstyle

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
	//Tool is an implementation of tool.Tool which allows
	//us to run Checkstyle on a Java class.
	Tool struct {
		java       string
		cmd        string
		configFile string
	}
)

//New
func New() *Tool {
	return &Tool{config.Config(config.JAVA),
		config.Config(config.CHECKSTYLE),
		config.Config(config.CHECKSTYLE_CONFIG)}
}

//Lang
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name
func (this *Tool) Name() string {
	return NAME
}

//Run runs checkstyle on the provided Java file. We make use of the configured Checkstyle configuration file.
//Output is written to an xml file which is then read in and used to create a Checkstyle Result.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "checkstyle.xml")
	args := []string{this.java, "-jar", this.cmd,
		"-f", "xml", "-c", this.configFile,
		"-o", outFile, "-r", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run checkstyle: %q.",
			string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
