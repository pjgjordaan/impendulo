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

package processor

import (
	"fmt"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	//P is used to process individual submissions.
	P interface {
		ResultName(tool.T) string
		Compile(bson.ObjectId, *tool.Target) error
		Tools() []tool.T
		Process(bson.ObjectId) error
	}
	FileP struct {
		sub      *project.Submission
		project  *project.Project
		rootDir  string
		srcDir   string
		toolDir  string
		jpfPath  string
		compiler tool.Compiler
		tools    []tool.T
	}
	TestP struct {
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

const (
	LOG_PROCESSOR = "processing/processor.go"
)

//NewFileP creates a P and sets up the environment and
//tools for it.
func NewFileP(sid bson.ObjectId) (*FileP, error) {
	s, e := db.Submission(bson.M{db.ID: sid}, nil)
	if e != nil {
		return nil, e
	}
	p, e := db.Project(bson.M{db.ID: s.ProjectId}, nil)
	if e != nil {
		return nil, e
	}
	d := filepath.Join(os.TempDir(), s.Id.Hex())
	fp := &FileP{
		sub:     s,
		project: p,
		rootDir: d,
		srcDir:  filepath.Join(d, "src"),
		toolDir: filepath.Join(d, "tools"),
	}
	//Can't proceed without our compiler
	fp.compiler, e = Compiler(fp)
	if e != nil {
		return nil, e
	}
	fp.tools, e = Tools(fp)
	if e != nil {
		return nil, e
	}
	return fp, nil
}

//Start listens for a submission's incoming files and processes them.
func (fp *FileP) Start(fc chan bson.ObjectId, dc chan util.E) {
	util.Log("Processing submission", fp.sub, LOG_PROCESSOR)
	//Processing loop.
processing:
	for {
		select {
		case fid := <-fc:
			if e := fp.Process(fid); e != nil {
				util.Log(e, LOG_PROCESSOR)
			}
			//Indicate that we are finished with the file.
			fc <- fid
		case <-dc:
			//We are done so time to exit.
			break processing
		}
	}
	if e := db.UpdateTime(fp.sub); e != nil {
		util.Log(e, LOG_PROCESSOR)
	}
	os.RemoveAll(fp.rootDir)
	util.Log("Processed submission", fp.sub, LOG_PROCESSOR)
	dc <- util.E{}
}

func (fp *FileP) Process(fid bson.ObjectId) error {
	//Retrieve file and process it.
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return e
	}
	switch f.Type {
	case project.TEST:
		return fp.Test(f)
	case project.ARCHIVE:
		return fp.Archive(f)
	case project.SRC:
		return fp.Source(f)
	default:
		return fmt.Errorf("cannot process file type %s", f.Type)
	}
}

//ProcessFile extracts archives and runs tools on source files.
func (fp *FileP) Source(f *project.File) error {
	util.Log("Processing file:", f.Id, LOG_PROCESSOR)
	defer util.Log("Processed file:", f.Id, LOG_PROCESSOR)
	//Create a target for the tools to run on and save the file.
	t := tool.NewTarget(f.Name, f.Package, fp.srcDir, tool.Language(fp.project.Lang))
	if e := util.SaveFile(t.FilePath(), f.Data); e != nil {
		return e
	}
	RunTools(f, t, fp)
	return nil
}

func (fp *FileP) Test(test *project.File) error {
	t := tool.NewTarget(test.Name, test.Package, fp.srcDir, tool.Language(fp.project.Lang))
	if e := util.SaveFile(t.FilePath(), test.Data); e != nil {
		return e
	}
	j, e := config.JUNIT.Path()
	if e != nil {
		return e
	}
	fp.compiler.AddCP(j)
	if e = fp.Compile(test.Id, t); e != nil {
		return e
	}
	tp, e := NewTestP(test, fp)
	if e != nil {
		return e
	}
	//we want to run the test on all of the submission's source files.
	fs, e := db.Files(bson.M{db.SUBID: test.SubId, db.TYPE: project.SRC}, bson.M{db.ID: 1}, 0)
	if e != nil {
		return e
	}
	for _, f := range fs {
		if e := tp.Process(f.Id); e != nil {
			util.Log(e, LOG_PROCESSOR)
		}
	}
	return nil
}

//Archive extracts files from an archive, stores and processes them.
func (fp *FileP) Archive(a *project.File) error {
	//Extract and store the files.
	m, e := util.UnzipToMap(a.Data)
	if e != nil {
		return e
	}
	for n, d := range m {
		if e = fp.archive(n, d); e != nil {
			util.Log(e, LOG_PROCESSOR)
		}
	}
	//We don't need the archive anymore
	return db.RemoveFileById(a.Id)
}

func (fp *FileP) archive(name string, data []byte) error {
	f, e := fp.StoreFile(name, data)
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
	e = fp.Process(f.Id)
	if r, e = request.RemoveFile(f); e != nil {
		return e
	}
	if se := mq.ChangeStatus(r); e == nil && se != nil {
		e = se
	}
	return e
}

//StoreFile creates a new project.File given an encoded file name and file data.
//The new project.File is then saved in the database.
func (fp *FileP) StoreFile(n string, d []byte) (*project.File, error) {
	f, e := project.ParseName(n)
	if e != nil {
		return nil, e
	}
	if db.Contains(db.FILES, bson.M{db.SUBID: fp.sub.Id, db.TYPE: f.Type, db.TIME: f.Time}) {
		return nil, db.DuplicateFile
	}
	f.SubId = fp.sub.Id
	f.Data = d
	if e := db.Add(db.FILES, f); e != nil {
		return nil, e
	}
	return f, nil
}

//Compile compiles a file, stores the result and returns any errors which may have occured.
func (fp *FileP) Compile(fid bson.ObjectId, t *tool.Target) error {
	r, e := fp.compiler.Run(fid, t)
	//We want to store the result if it is a compilation error
	if e == nil || tool.IsCompileError(e) {
		db.AddResult(r, fp.compiler.Name())
	}
	return e
}

func (fp *FileP) ResultName(t tool.T) string {
	return t.Name()
}

func (fp *FileP) Tools() []tool.T {
	return fp.tools
}

func NewTestP(tf *project.File, fp *FileP) (*TestP, error) {
	d := filepath.Join(fp.rootDir, tf.Id.Hex())
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
	tp := &TestP{
		sub:      fp.sub,
		project:  fp.project,
		id:       tf.Id,
		rootDir:  d,
		srcDir:   srcDir,
		toolDir:  td,
		compiler: c,
	}
	tp.tools, e = TestTools(tp, tf)
	if e != nil {
		return nil, e
	}
	return tp, nil
}

func (tp *TestP) Process(fid bson.ObjectId) error {
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return e
	}
	t := tool.NewTarget(f.Name, f.Package, tp.srcDir, tool.JAVA)
	if e = util.SaveFile(t.FilePath(), f.Data); e != nil {
		return e
	}
	return RunTools(f, t, tp)
}

func (tp *TestP) ResultName(t tool.T) string {
	return t.Name() + "-" + tp.id.Hex()
}

func (tp *TestP) Compile(fid bson.ObjectId, t *tool.Target) error {
	_, e := tp.compiler.Run(fid, t)
	return e
}

func (tp *TestP) Tools() []tool.T {
	return tp.tools
}

//RunTools runs all available tools on a file. It skips a tool if
//there is already a result for it present. This makes it possible to
//rerun old tools or add new tools and run them on old files without having
//to rerun all the tools.
func RunTools(file *project.File, target *tool.Target, p P) error {
	if e := p.Compile(file.Id, target); e != nil {
		return e
	}
	for _, t := range p.Tools() {
		if e := RunTool(t, file, target, p.ResultName(t)); e != nil {
			util.Log(e, LOG_PROCESSOR)
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
	//tf, _ := db.File(bson.M{db.ID: f.Id}, bson.M{db.RESULTS: 1})
	//fmt.Println(tf.Results)
	return de
}
