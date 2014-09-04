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

package worker

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Worker is used to process individual submissions.
	W interface {
		ResultName(tool.T) string
		Compile(bson.ObjectId, *tool.Target) error
		Tools() []tool.T
		Process(bson.ObjectId) error
	}
)

const (
	LOG_W = "processing/worker/worker.go"
)

//RunTools runs all available tools on a file. It skips a tool if
//there is already a result for it present. This makes it possible to
//rerun old tools or add new tools and run them on old files without having
//to rerun all the tools.
func RunTools(file *project.File, target *tool.Target, w W) error {
	if e := w.Compile(file.Id, target); e != nil {
		return e
	}
	for _, t := range w.Tools() {
		if e := RunTool(t, file, target, w.ResultName(t)); e != nil {
			util.Log(e, LOG_W)
		}
	}
	return nil
}

//RunTool runs a specified tool on a file. The target specifies where the file is stored.
func RunTool(t tool.T, f *project.File, target *tool.Target, n string) error {
	if _, ok := f.Results[n]; ok {
		return nil
	}
	var de error
	r, e := t.Run(f.Id, target)
	if e != nil {
		//Report any errors and store timeouts.
		if tool.IsTimeout(e) {
			de = db.AddFileResult(f.Id, n, result.TIMEOUT)
		} else {
			de = db.AddFileResult(f.Id, n, result.ERROR)
		}
	} else if r != nil {
		de = db.AddResult(r, n)
	}
	if e == nil && de != nil {
		e = de
	}
	return de
}
