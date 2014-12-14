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

package file

import (
	"container/list"
	"fmt"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/worker"
	"github.com/godfried/impendulo/processor/worker/test"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	Worker struct {
		submission  *project.Submission
		project     *project.P
		rootDir     string
		srcDir      string
		toolDir     string
		jpwath      string
		compiler    tool.Compiler
		tools       []tool.T
		filer       *mq.MessageHandler
		requestChan chan *request.R
	}
)

const (
	LOG_F = "processing/file/file.go"
)

//New creates a P and sets up the environment and
//tools for it.
func New(sid bson.ObjectId) (*Worker, error) {
	s, e := db.Submission(bson.M{db.ID: sid}, nil)
	if e != nil {
		return nil, e
	}
	p, e := db.Project(bson.M{db.ID: s.ProjectId}, nil)
	if e != nil {
		return nil, e
	}
	d := filepath.Join(os.TempDir(), s.Id.Hex())
	rc := make(chan *request.R)
	f, e := mq.NewFiler(rc, sid)
	if e != nil {
		return nil, e
	}
	w := &Worker{
		submission:  s,
		project:     p,
		rootDir:     d,
		srcDir:      filepath.Join(d, "src"),
		toolDir:     filepath.Join(d, "tools"),
		filer:       f,
		requestChan: rc,
	}
	//Can't proceed without our compiler
	w.compiler, e = Compiler(w)
	if e != nil {
		return nil, e
	}
	w.tools, e = Tools(w)
	if e != nil {
		return nil, e
	}
	return w, nil
}

//Start listens for a submission's incoming files and processes them.
func (w *Worker) Start(done chan util.E) {
	defer w.shutdown(done)
	go mq.H(w.filer)
	w.processFiles()
	if e := db.UpdateTime(w.submission); e != nil {
		util.Log(e, LOG_F)
	}
	os.RemoveAll(w.rootDir)
}

func (w *Worker) shutdown(done chan util.E) {
	done <- util.E{}
	if e := mq.ChangeStatus(request.StopSubmission(w.submission.Id)); e != nil {
		util.Log(e, LOG_F)
	}
	w.filer.Shutdown()
}

func (w *Worker) processFiles() {
	tq := list.New()
	for {
		if w.handleRequest(tq, <-w.requestChan) {
			break
		}
	}
	for e := tq.Front(); e != nil; e = e.Next() {
		if e := w.Process(e.Value.(bson.ObjectId)); e != nil {
			util.Log(e, LOG_F)
		}
	}
}

func (w *Worker) handleRequest(tq *list.List, r *request.R) bool {
	switch r.Type {
	case request.SUBMISSION_STOP:
		return true
	case request.SRC_ADD, request.ARCHIVE_ADD:
		if e := w.Process(r.FileId); e != nil {
			util.Log(e, LOG_F)
		}
	case request.TEST_ADD:
		tq.PushBack(r.FileId)
	default:
		util.Log(fmt.Errorf("unsupported request %s", r), LOG_F)
	}
	return false
}

func (w *Worker) Process(fid bson.ObjectId) error {
	e := w.process(fid)
	e2 := removeFile(fid)
	if e == nil && e2 != nil {
		e = e2
	}
	return e
}

func (w *Worker) process(fid bson.ObjectId) error {
	//Retrieve file and process it.
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return e
	}
	switch f.Type {
	case project.TEST:
		return w.Test(f)
	case project.ARCHIVE:
		return w.Archive(f)
	case project.SRC:
		return w.Source(f)
	default:
		return fmt.Errorf("cannot process file type %s", f.Type)
	}
}

//ProcessFile extracts archives and runs tools on source files.
func (w *Worker) Source(f *project.File) error {
	//Create a target for the tools to run on and save the file.
	t := tool.NewTarget(f.Name, f.Package, w.srcDir, project.Language(w.project.Lang))
	if e := util.SaveFile(t.FilePath(), f.Data); e != nil {
		return e
	}
	worker.RunTools(f, t, w)
	return nil
}

func (w *Worker) Test(tf *project.File) error {
	t := tool.NewTarget(tf.Name, tf.Package, w.srcDir, project.Language(w.project.Lang))
	if e := util.SaveFile(t.FilePath(), tf.Data); e != nil {
		return e
	}
	j, e := config.JUNIT.Path()
	if e != nil {
		return e
	}
	w.compiler.AddCP(j)
	if e = w.Compile(tf.Id, t); e != nil {
		return e
	}
	tp, e := test.New(tf, w.submission, w.project, w.rootDir)
	if e != nil {
		return e
	}
	//we want to run the test on all of the submission's source files.
	fs, e := db.Files(bson.M{db.SUBID: tf.SubId, db.TYPE: project.SRC}, bson.M{db.ID: 1}, 0)
	if e != nil {
		return e
	}
	for _, f := range fs {
		if e := tp.Process(f.Id); e != nil {
			util.Log(e, LOG_F)
		}
	}
	return nil
}

//Archive extracts files from an archive, stores and processes them.
func (w *Worker) Archive(a *project.File) error {
	//Extract and store the files.
	m, e := util.UnzipToMap(a.Data)
	if e != nil {
		return e
	}
	for n, d := range m {
		if e = w.archive(n, d); e != nil {
			util.Log(e, LOG_F)
		}
	}
	return nil
}

func (w *Worker) archive(name string, data []byte) error {
	f, e := w.StoreFile(name, data)
	if e != nil {
		return e
	}
	r, e := request.AddFile(f)
	if e != nil {
		return e
	}
	if e := mq.ChangeStatus(r); e != nil {
		return e
	}
	return w.Process(f.Id)
}

//StoreFile creates a new project.File given an encoded file name and file data.
//The new project.File is then saved in the database.
func (w *Worker) StoreFile(n string, d []byte) (*project.File, error) {
	f, e := project.ParseFile(n, d)
	if e != nil {
		return nil, e
	}
	if db.Contains(db.FILES, bson.M{db.SUBID: w.submission.Id, db.TYPE: f.Type, db.TIME: f.Time}) {
		return nil, db.DuplicateFile
	}
	f.SubId = w.submission.Id
	if e := db.Add(db.FILES, f); e != nil {
		return nil, e
	}
	return f, nil
}

//Compile compiles a file, stores the result and returns any errors which may have occured.
func (w *Worker) Compile(fid bson.ObjectId, t *tool.Target) error {
	r, e := w.compiler.Run(fid, t)
	//We want to store the result if it is a compilation error
	if e == nil || tool.IsCompileError(e) {
		db.AddResult(r, w.compiler.Name())
	}
	return e
}

func (w *Worker) ResultName(t tool.T) string {
	return t.Name()
}

func (w *Worker) Tools() []tool.T {
	return w.tools
}

func removeFile(fid bson.ObjectId) error {
	f, e := db.File(bson.M{db.ID: fid}, db.FILE_SELECTOR)
	if e != nil {
		return e
	}
	r, e := request.RemoveFile(f)
	if e != nil {
		return e
	}
	if e = mq.ChangeStatus(r); e != nil {
		return e
	}
	if r.Type != request.ARCHIVE_REMOVE {
		return nil
	}
	return db.RemoveFileById(fid)
}
