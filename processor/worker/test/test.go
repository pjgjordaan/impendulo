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

package test

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/worker"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	Worker struct {
		sub      *project.Submission
		project  *project.Project
		id       bson.ObjectId
		rootDir  string
		srcDir   string
		toolDir  string
		compiler tool.Compiler
		tools    []tool.T
	}
)

func New(f *project.File, s *project.Submission, p *project.Project, rootDir string) (*Worker, error) {
	d := filepath.Join(rootDir, f.Id.Hex())
	td := filepath.Join(d, "tools")
	c, e := javac.New("")
	if e != nil {
		return nil, e
	}
	rd, e := config.JUNIT_TESTING.Path()
	if e != nil {
		return nil, e
	}
	if e = util.Copy(td, rd); e != nil {
		return nil, e
	}
	srcDir := filepath.Join(d, "src")
	if e = os.MkdirAll(srcDir, util.DPERM); e != nil {
		return nil, e
	}
	w := &Worker{
		sub:      s,
		project:  p,
		id:       f.Id,
		rootDir:  d,
		srcDir:   srcDir,
		toolDir:  td,
		compiler: c,
	}
	w.tools, e = TestTools(w, f)
	if e != nil {
		return nil, e
	}
	return w, nil
}

func (w *Worker) Process(fid bson.ObjectId) error {
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return e
	}
	t := tool.NewTarget(f.Name, f.Package, w.srcDir, tool.JAVA)
	if e = util.SaveFile(t.FilePath(), f.Data); e != nil {
		return e
	}
	return worker.RunTools(f, t, w)
}

func (w *Worker) ResultName(t tool.T) string {
	return t.Name() + "-" + w.id.Hex()
}

func (w *Worker) Compile(fid bson.ObjectId, t *tool.Target) error {
	_, e := w.compiler.Run(fid, t)
	return e
}

func (w *Worker) Tools() []tool.T {
	return w.tools
}
