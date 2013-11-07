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

//Package checkstyle is the Checkstyle static analysis tool's implementation of an Impendulo tool.
//See http://checkstyle.sourceforge.net/ for more information.
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
		java string
		cmd  string
		cfg  string
	}
)

//New creates a new instance of the checkstyle Tool.
//Any errors returned will of type config.ConfigError.
func New() (tool *Tool, err error) {
	tool = new(Tool)
	tool.java, err = config.JAVA.Path()
	if err != nil {
		return
	}
	tool.cmd, err = config.CHECKSTYLE.Path()
	if err != nil {
		return
	}
	tool.cfg, err = config.CHECKSTYLE_CFG.Path()
	return
}

//Lang is Java
func (this *Tool) Lang() tool.Language {
	return tool.JAVA
}

//Name is Checkstyle
func (this *Tool) Name() string {
	return NAME
}

//Run runs checkstyle on the provided Java file. We make use of the configured Checkstyle configuration file.
//Output is written to an xml file which is then read in and used to create a Checkstyle Result.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "checkstyle.xml")
	args := []string{this.java, "-jar", this.cmd,
		"-f", "xml", "-c", this.cfg,
		"-o", outFile, "-r", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
		if err != nil && execRes.Err != nil {
			err = execRes.Err
		}
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run checkstyle: %q.",
			string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
