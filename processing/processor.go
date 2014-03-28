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

package processing

import (
	"fmt"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	//Processor is used to process individual submissions.
	Processor struct {
		sub      *project.Submission
		project  *project.Project
		rootDir  string
		srcDir   string
		toolDir  string
		jpfPath  string
		compiler tool.Tool
		tools    []tool.Tool
	}
	TestProcessor struct {
		sub      *project.Submission
		project  *project.Project
		id       bson.ObjectId
		results  bson.M
		rootDir  string
		srcDir   string
		toolDir  string
		compiler tool.Tool
		tools    []tool.Tool
	}
)

const (
	LOG_PROCESSOR = "processing/processor.go"
)

//NewProcessor creates a Processor and sets up the environment and
//tools for it.
func NewProcessor(sid bson.ObjectId) (*Processor, error) {
	s, e := db.Submission(bson.M{db.ID: sid}, nil)
	if e != nil {
		return nil, e
	}
	p, e := db.Project(bson.M{db.ID: s.ProjectId}, nil)
	if e != nil {
		return nil, e
	}
	d := filepath.Join(os.TempDir(), s.Id.Hex())
	proc := &Processor{
		sub:     s,
		project: p,
		rootDir: d,
		srcDir:  filepath.Join(d, "src"),
		toolDir: filepath.Join(d, "tools"),
	}
	//Can't proceed without our compiler
	proc.compiler, e = Compiler(proc)
	if e != nil {
		return nil, e
	}
	proc.tools, e = Tools(proc)
	if e != nil {
		return nil, e
	}
	return proc, nil
}

//Process listens for a submission's incoming files and processes them.
func (p *Processor) Process(fc chan bson.ObjectId, dc chan E) {
	util.Log("Processing submission", p.sub, LOG_PROCESSOR)
	defer func() {
		os.RemoveAll(p.rootDir)
		util.Log("Processed submission", p.sub, LOG_PROCESSOR)
		dc <- E{}
	}()
	//Processing loop.
processing:
	for {
		select {
		case fId := <-fc:
			if e := p.ProcessFile(fId); e != nil {
				util.Log(e, LOG_PROCESSOR)
			}
			//Indicate that we are finished with the file.
			fc <- fId
		case <-dc:
			//We are done so time to exit.
			break processing
		}
	}
	if e := db.UpdateStatus(p.sub); e != nil {
		util.Log(e, LOG_PROCESSOR)
	}
	if e := db.UpdateTime(p.sub); e != nil {
		util.Log(e, LOG_PROCESSOR)
	}
}

func (p *Processor) ProcessFile(fid bson.ObjectId) error {
	//Retrieve file and process it.
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return e
	}
	switch f.Type {
	case project.TEST:
		return p.ProcessTest(f)
	case project.ARCHIVE:
		return p.ProcessArchive(f)
	case project.SRC:
		return p.ProcessSource(f)
	}
	return fmt.Errorf("cannot process file type %s", f.Type)
}

//ProcessFile extracts archives and runs tools on source files.
func (p *Processor) ProcessSource(f *project.File) error {
	util.Log("Processing file:", f, LOG_PROCESSOR)
	defer util.Log("Processed file:", f, LOG_PROCESSOR)
	//Create a target for the tools to run on and save the file.
	t := tool.NewTarget(f.Name, f.Package, p.srcDir, tool.Language(p.project.Lang))
	if e := util.SaveFile(t.FilePath(), f.Data); e != nil {
		util.Log(e, LOG_PROCESSOR)
		return e
	}
	p.RunTools(f, t)
	return nil
}

func (p *Processor) ProcessTest(test *project.File) error {
	t := tool.NewTarget(test.Name, test.Package, p.srcDir, tool.Language(p.project.Lang))
	if e := util.SaveFile(t.FilePath(), test.Data); e != nil {
		return e
	}
	j, e := config.JUNIT.Path()
	if e != nil {
		return e
	}
	p.compiler.(*javac.Tool).AddCP(j)
	if e = p.Compile(test.Id, t); e != nil {
		return e
	}
	tp, e := NewTestProcessor(test, p)
	if e != nil {
		return e
	}
	fs, e := db.Files(bson.M{db.SUBID: test.SubId, db.TYPE: project.SRC}, bson.M{db.ID: 1})
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

//Extract extracts files from an archive, stores and processes them.
func (p *Processor) ProcessArchive(a *project.File) error {
	//Extract and store the files.
	m, e := util.UnzipToMap(a.Data)
	if e != nil {
		return e
	}
	for n, d := range m {
		if e = p.StoreFile(n, d); e != nil {
			util.Log(e, LOG_PROCESSOR)
		}
	}
	//We don't need the archive anymore
	if e = db.RemoveFileById(a.Id); e != nil {
		util.Log(e, LOG_PROCESSOR)
	}
	fs, e := db.Files(bson.M{db.SUBID: p.sub.Id}, bson.M{db.TIME: 1, db.ID: 1}, db.TIME)
	if e != nil {
		return e
	}
	ChangeStatus(Status{len(fs), 0})
	//Process archive files.
	for _, f := range fs {
		if e = p.ProcessFile(f.Id); e != nil {
			util.Log(e, LOG_PROCESSOR)
		}
		ChangeStatus(Status{-1, 0})
	}
	return nil
}

//StoreFile creates a new project.File given an encoded file name and file data.
//The new project.File is then saved in the database.
func (p *Processor) StoreFile(n string, d []byte) error {
	f, e := project.ParseName(n)
	if e != nil {
		return e
	}
	if db.Contains(db.FILES, bson.M{db.SUBID: p.sub.Id, db.TYPE: f.Type, db.TIME: f.Time}) {
		return nil
	}
	f.SubId = p.sub.Id
	f.Data = d
	return db.Add(db.FILES, f)
}

//RunTools runs all available tools on a file. It skips a tool if
//there is already a result for it present. This makes it possible to
//rerun old tools or add new tools and run them on old files without having
//to rerun all the tools.
func (p *Processor) RunTools(f *project.File, target *tool.Target) {
	//First we compile our file.
	if e := p.Compile(f.Id, target); e != nil {
		util.Log(e, LOG_PROCESSOR)
		return
	}
	for _, t := range p.tools {
		//Skip the tool if it has already been run.
		if _, ok := f.Results[t.Name()]; ok {
			continue
		}
		r, e := t.Run(f.Id, target)
		if e != nil {
			//Report any errors and store timeouts.
			util.Log(fmt.Errorf("error %q running tool %s on file %s.", e, t.Name(), f.Id.Hex()), LOG_PROCESSOR)
			if tool.IsTimeout(e) {
				e = db.AddFileResult(f.Id, t.Name(), tool.TIMEOUT)
			} else {
				e = db.AddFileResult(f.Id, t.Name(), tool.ERROR)
			}
		} else if r != nil {
			e = db.AddResult(r, t.Name())
		} else {
			e = db.AddFileResult(f.Id, t.Name(), tool.NORESULT)
		}
		if e != nil {
			util.Log(e, LOG_PROCESSOR)
		}
	}
}

//Compile compiles a file, stores the result and returns any errors which may have occured.
func (p *Processor) Compile(fid bson.ObjectId, t *tool.Target) error {
	r, e := p.compiler.Run(fid, t)
	//We want to store the result if it is a compilation error
	if e == nil || tool.IsCompileError(e) {
		db.AddResult(r, p.compiler.Name())
	}
	return e
}

func NewTestProcessor(tf *project.File, p *Processor) (*TestProcessor, error) {
	d := filepath.Join(p.rootDir, tf.Id.Hex())
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
	tp := &TestProcessor{
		sub:      p.sub,
		project:  p.project,
		id:       tf.Id,
		results:  tf.Results,
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

func (tp *TestProcessor) Process(fid bson.ObjectId) error {
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return e
	}
	target := tool.NewTarget(f.Name, f.Package, tp.srcDir, tool.JAVA)
	if e = util.SaveFile(target.FilePath(), f.Data); e != nil {
		return e
	}
	if _, e = tp.compiler.Run(fid, target); e != nil {
		return e
	}
	for _, c := range tp.tools {
		if e := tp.Run(c, f, target); e != nil {
			util.Log(e, LOG_PROCESSOR)
		}
	}
	return nil
}

func (tp *TestProcessor) Run(t tool.Tool, f *project.File, target *tool.Target) error {
	n := t.Name() + "-" + f.Id.Hex()
	if _, ok := tp.results[n]; ok {
		return nil
	}
	r, e := t.Run(tp.id, target)
	if e != nil {
		util.Log(e, LOG_PROCESSOR)
		//Report any errors and store timeouts.
		if tool.IsTimeout(e) {
			return db.AddFileResult(tp.id, n, tool.TIMEOUT)
		} else {
			return db.AddFileResult(tp.id, n, tool.ERROR)
		}
	} else if r != nil {
		return db.AddResult(r, n)
	}
	return db.AddFileResult(tp.id, n, tool.NORESULT)
}
