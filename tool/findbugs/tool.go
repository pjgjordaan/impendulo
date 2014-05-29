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

//Package findbugs is the Findbugs static analysis tool's implementation of an Imendulo tool.
//See http://findbugs.sourceforge.net/ for more information.
package findbugs

import (
	"fmt"
	"time"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	//Findbugs is a tool.Tool used to run Findbugs on Java classes.
	Tool struct {
		cmd string
	}
)

//New creates a new instance of the Findbugs tool.
//If an error is returned, it will be due Findbugs not being configured correctly.
func New() (*Tool, error) {
	p, e := config.FINDBUGS.Path()
	if e != nil {
		return nil, e
	}
	return &Tool{cmd: p}, nil
}

//Lang is Java.
func (t *Tool) Lang() tool.Language {
	return tool.JAVA
}

//Name is Findbugs.
func (t *Tool) Name() string {
	return NAME
}

//Run executes Findbugs on the provided source file.
//Findbugs is run with the following flags: -effort:max, -experimental, -relaxed.
//The result is written to an XML file which is then read and used to create a
//Findbugs Result.
func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (tool.ToolResult, error) {
	//Setup arguments
	jp, e := config.JAVA.Path()
	if e != nil {
		return nil, e
	}
	o := filepath.Join(target.Dir, "findbugs.xml")
	a := []string{jp, "-jar", t.cmd, "-textui", "-effort:max", "-experimental",
		"-xml:withMessages", "-relaxed", "-output", o, target.PackagePath()}
	defer os.Remove(o)
	//Run Findbugs and load result.
	r, re := tool.RunCommand(a, nil, 30*time.Second)
	rf, e := os.Open(o)
	if e != nil {
		if re != nil {
			return nil, re
		}
		return nil, fmt.Errorf("could not run findbugs: %q", string(r.StdErr))
	}
	nr, e := NewResult(fileId, util.ReadBytes(rf))
	if e != nil {
		if re != nil {
			e = re
		}
		return nil, e
	}
	return nr, nil
}
