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
	"time"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
	"strings"
)

type (
	//Tool is an implementation of tool.T which allows us to run
	//PMD on Java classes.
	Tool struct {
		cmd   string
		rules string
	}
)

//New creates a new instance of a PMD Tool.
//Errors returned will be due to loading either the default
//PMD rules or the PMD execution script.
func New(rules *Rules) (*Tool, error) {
	if rules == nil {
		var e error
		rules, e = DefaultRules(bson.NewObjectId())
		if e != nil {
			return nil, e
		}
	}
	p, e := config.PMD.Path()
	if e != nil {
		return nil, e
	}
	return &Tool{
		cmd:   p,
		rules: strings.Join(rules.RuleArray(), ","),
	}, nil
}

//Lang is Java
func (t *Tool) Lang() project.Language {
	return project.JAVA
}

//Name is PMD
func (t *Tool) Name() string {
	return NAME
}

//Run runs PMD on a provided Java source file. PMD writes its output to an XML file which we then read
//and use to create a PMD Result.
func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (result.Tooler, error) {
	o := filepath.Join(target.Dir, "pmd.xml")
	a := []string{t.cmd, "pmd", "-f", "xml", "-stress", "-shortnames", "-R", t.rules, "-r", o, "-d", target.Dir}
	defer os.Remove(o)
	r, re := tool.RunCommand(a, nil, 30*time.Second)
	rf, e := os.Open(o)
	if e != nil {
		if re != nil {
			return nil, re
		}
		return nil, fmt.Errorf("could not run pmd: %s", string(r.StdErr))
	}
	nr, e := NewResult(fileId, util.ReadBytes(rf))
	if e == nil {
		return nr, nil
	}
	if re != nil {
		e = re
	}
	return nil, e
}
