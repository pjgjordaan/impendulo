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
func New(cfg *Config, jpfDir string) (*Tool, error) {
	//Load locations
	rd, e := config.JPF_RUNNER.Path()
	if e != nil {
		return nil, e
	}
	jar, e := config.JPF.Path()
	if e != nil {
		return nil, e
	}
	rjar, e := config.JPF_RUN.Path()
	if e != nil {
		return nil, e
	}
	g, e := config.GSON.Path()
	if e != nil {
		return nil, e
	}
	//Copy JPF runner files to the specified location
	if e = util.Copy(jpfDir, rd); e != nil {
		return nil, e
	}
	//Save the config file in the same location.
	jp := filepath.Join(jpfDir, "config.jpf")
	if e = util.SaveFile(jp, cfg.Data); e != nil {
		return nil, e
	}
	//Setup classpath with the required JPF and Json jars.
	cp := jpfDir + ":" + jar + ":" + rjar + ":" + g
	//Compile JPF runner files
	jt := tool.NewTarget("JPFRunner.java", "runner", jpfDir, tool.JAVA)
	pt := tool.NewTarget("ImpenduloPublisher.java", "runner", jpfDir, tool.JAVA)
	c, e := javac.New(cp)
	if e != nil {
		return nil, e
	}
	id := bson.NewObjectId()
	if _, e = c.Run(id, jt); e != nil {
		return nil, e
	}
	if _, e = c.Run(id, pt); e != nil {
		return nil, e
	}
	return &Tool{cp: cp, jpfPath: jp, exec: jt.Executable()}, nil
}

//Lang is Java
func (t *Tool) Lang() tool.Language {
	return tool.JAVA
}

//Name is JPF
func (t *Tool) Name() string {
	return NAME
}

//Run runs JPF on a specified Java source file. It uses the Java class runner.JPFRunner to
//actually run JPF on the source file. If the command was successful, the
//results are read in from a xml file.
func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (tool.ToolResult, error) {
	//Load arguments
	jp, e := config.JAVA.Path()
	if e != nil {
		return nil, e
	}
	o := filepath.Join(target.Dir, "jpf")
	a := []string{jp, "-cp", target.Dir + ":" + t.cp, t.exec, t.jpfPath, target.Executable(), target.Dir, o}
	o = o + ".xml"
	defer os.Remove(o)
	//Run JPF and load result
	r, re := tool.RunCommand(a, nil)
	rf, e := os.Open(o)
	if e == nil {
		//Tests ran successfully.
		nr, e := NewResult(fileId, util.ReadBytes(rf))
		if e != nil {
			if re != nil {
				e = re
			}
			return nil, e
		}
		return nr, nil
	} else if r.HasStdErr() {
		return nil, fmt.Errorf("could not execute jpf runner: %q", string(r.StdErr))
	}
	return nil, re
}
