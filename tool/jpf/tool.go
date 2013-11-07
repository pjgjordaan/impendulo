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

//Package jpf is the Java Pathfinder verification system's implementation of an Impendulo tool.
//See http://babelfish.arc.nasa.gov/trac/jpf/ for more information.
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
	//Load locations
	runDir, err := config.JPF_RUNNER.Path()
	if err != nil {
		return
	}
	jpfJar, err := config.JPF.Path()
	if err != nil {
		return
	}
	jpfRunJar, err := config.JPF_RUN.Path()
	if err != nil {
		return
	}
	gson, err := config.GSON.Path()
	if err != nil {
		return
	}
	//Copy JPF runner files to the specified location
	err = util.Copy(jpfDir, runDir)
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
	cp := jpfDir + ":" + jpfJar + ":" + jpfRunJar + ":" + gson
	//Compile JPF runner files
	jpfInfo := tool.NewTarget("JPFRunner.java", "runner", jpfDir, tool.JAVA)
	pubInfo := tool.NewTarget("ImpenduloPublisher.java", "runner", jpfDir, tool.JAVA)
	comp, err := javac.New(cp)
	if err != nil {
		return
	}
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
func (this *Tool) Lang() tool.Language {
	return tool.JAVA
}

//Name is JPF
func (this *Tool) Name() string {
	return NAME
}

//Run runs JPF on a specified Java source file. It uses the Java class runner.JPFRunner to
//actually run JPF on the source file. If the command was successful, the
//results are read in from a xml file.
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	//Load arguments
	java, err := config.JAVA.Path()
	if err != nil {
		return
	}
	outFile := filepath.Join(ti.Dir, "jpf")
	args := []string{java, "-cp", ti.Dir + ":" + this.cp, this.exec,
		this.jpfPath, ti.Executable(), ti.Dir, outFile}
	outFile = outFile + ".xml"
	defer os.Remove(outFile)
	//Run JPF and load result
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
		err = fmt.Errorf("Could not execute jpf runner: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}
