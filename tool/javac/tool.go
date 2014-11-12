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

//Package javac is the OpenJDK Java compiler's implementation of an Impendulo tool.
//For more information see http://openjdk.java.net/groups/compiler/.
package javac

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"

	"time"
)

type (
	Tool struct {
		cmd string
		cp  string
	}
)

//New creates a new javac instance. cp is the classpath used when compiling.
func New(cp string) (*Tool, error) {
	p, e := config.JAVAC.Path()
	if e != nil {
		return nil, e
	}
	return &Tool{cp: cp, cmd: p}, nil
}

//Lang is Java.
func (t *Tool) Lang() project.Language {
	return project.JAVA
}

//Name is Javac
func (t *Tool) Name() string {
	return NAME
}

func (t *Tool) AddCP(s string) {
	if t.cp != "" {
		t.cp += ":"
	}
	t.cp += s
}

//Run compiles the Java source file specified by t. We compile with maximum warnings and compile
//classes implicitly loaded by the source code. All compilation results will be stored (success,
//errors and warnings).
func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (result.Tooler, error) {
	cp := t.cp
	if cp != "" {
		cp += ":"
	}
	cp += target.Dir
	a := []string{t.cmd, "-cp", cp + ":" + target.Dir, "-implicit:class", "-Xlint", target.FilePath()}
	r, e := tool.RunCommand(a, nil, 30*time.Second)
	if e != nil {
		if !tool.IsEndError(e) {
			return nil, e
		}
		return NewResult(fileId, r.StdErr), tool.NewCompileError(target.FullName(), string(r.StdErr))
	} else if r.HasStdErr() {
		//Compiler warnings.
		return NewResult(fileId, r.StdErr), nil
	}
	return NewResult(fileId, result.COMPILE_SUCCESS), nil
}
