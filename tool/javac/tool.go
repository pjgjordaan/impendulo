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

//Package javac is the OpenJDK Java compiler's implementation of an Impendulo tool.
//For more information see http://openjdk.java.net/groups/compiler/.
package javac

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type (
	//Javac is a tool.Tool used to compile Java source files.
	Tool struct {
		cmd string
		cp  string
	}
)

//New creates a new javac instance. cp is the classpath used when compiling.
func New(cp string) (tool *Tool, err error) {
	tool = &Tool{
		cp: cp,
	}
	tool.cmd, err = config.Binary(config.JAVAC)
	return
}

//Lang is Java.
func (this *Tool) Lang() string {
	return tool.JAVA
}

//Name is Javac
func (this *Tool) Name() string {
	return NAME
}

//Run compiles the Java source file specified by ti. We compile with maximum warnings and compile
//classes implicitly loaded by the source code. All compilation results will be stored (success,
//errors and warnings).
func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	if this.cp != "" {
		this.cp += ":"
	}
	this.cp += ti.Dir
	args := []string{this.cmd, "-cp", this.cp + ":" + ti.Dir,
		"-implicit:class", "-Xlint", ti.FilePath()}
	//Compile the file.
	execRes := tool.RunCommand(args, nil)
	if execRes.Err != nil {
		if !tool.IsEndError(execRes.Err) {
			err = execRes.Err
		} else {
			//Unsuccessfull compile.
			res = NewResult(fileId, execRes.StdErr)
			err = tool.NewCompileError(ti.FullName(), string(execRes.StdErr))
		}
	} else if execRes.HasStdErr() {
		//Compiler warnings.
		res = NewResult(fileId, execRes.StdErr)
	} else {
		res = NewResult(fileId, compSuccess)
	}
	return
}
