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

//Package jpf is Java Pathfinder's implementation of an Impendulo tool.
package jpf

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
	//Tool is an implementation of tool.Tool which runs a JPF on a Java source file.
	//It makes use of runner.JPFRunner to configure and run JPF on a file and
	//runner.ImpenduloPublisher to output the results as XML.
	Tool struct {
		cp, jpfPath, exec string
	}
)

//New creates a new JPF instance for a given submission. jpfDir is where the
//Java JPF runner files should be stored for this JPF instance.
//jpfConfig is the JPF configuration associated with the submission's project.
func New(jpfConfig *Config, jpfDir string) (jpf *Tool, err error) {
	//Copy JPF runner files to the specified location
	err = util.Copy(jpfDir, config.Config(config.JPF_RUNNER_DIR))
	if err != nil {
		return
	}
	//Save the config file in the same location.
	jpfPath := filepath.Join(jpfDir, "config.jpf")
	err = util.SaveFile(jpfPath, jpfConfig.Data)
	if err != nil {
		return
	}
	//Setup classpath with the required JPF and Json jars.
	cp := jpfDir + ":" + config.Config(config.JPF_JAR) + ":" +
		config.Config(config.RUNJPF_JAR) + ":" + config.Config(config.GSON_JAR)
	//Compile JPF runner files
	jpfInfo := tool.NewTarget("JPFRunner.java", "java", "runner", jpfDir)
	pubInfo := tool.NewTarget("ImpenduloPublisher.java", "java", "runner", jpfDir)
	comp := javac.New(cp)
	id := bson.NewObjectId()
	_, err = comp.Run(id, jpfInfo)
	if err != nil {
		return
	}
	_, err = comp.Run(id, pubInfo)
	if err != nil {
		return
	}
	jpf = &Tool{
		cp:      cp,
		jpfPath: jpfPath,
		exec:    jpfInfo.Executable(),
	}
	return
}

//Lang is Java
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name is JPF
func (this *Tool) Name() string {
	return NAME
}

//Run runs JPF on a specified Java source file. It uses runner.JPFRunner to actually run JPF
//on the source file. If the command was successful, the results are read in from a xml file.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "jpf")
	args := []string{config.Config(config.JAVA), "-cp", ti.Dir + ":" +
		this.cp, this.exec, this.jpfPath, ti.Executable(),
		ti.Dir, outFile}
	outFile = outFile + ".xml"
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not execute jpf runner: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
