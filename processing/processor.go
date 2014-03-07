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
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/junit"
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
func NewProcessor(subId bson.ObjectId) (proc *Processor, err error) {
	sub, err := db.Submission(bson.M{db.ID: subId}, nil)
	if err != nil {
		return
	}
	matcher := bson.M{db.ID: sub.ProjectId}
	proj, err := db.Project(matcher, nil)
	if err != nil {
		return
	}
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	toolDir := filepath.Join(dir, "tools")
	proc = &Processor{
		sub:     sub,
		project: proj,
		rootDir: dir,
		srcDir:  filepath.Join(dir, "src"),
		toolDir: toolDir,
	}
	//Can't proceed without our compiler
	proc.compiler, err = Compiler(proc)
	if err != nil {
		return
	}
	proc.tools, err = Tools(proc)
	return
}

//Process listens for a submission's incoming files and processes them.
func (this *Processor) Process(fileChan chan bson.ObjectId, doneChan chan E) {
	util.Log("Processing submission", this.sub, LOG_PROCESSOR)
	defer os.RemoveAll(this.rootDir)
	//Processing loop.
processing:
	for {
		select {
		case fId := <-fileChan:
			err := this.ProcessFile(fId)
			if err != nil {
				util.Log(err, LOG_PROCESSOR)
			}
			//Indicate that we are finished with the file.
			fileChan <- fId
		case <-doneChan:
			//We are done so time to exit.
			break processing
		}
	}
	err := db.UpdateStatus(this.sub)
	if err != nil {
		util.Log(err, LOG_PROCESSOR)
	}
	err = db.UpdateTime(this.sub)
	if err != nil {
		util.Log(err, LOG_PROCESSOR)
	}
	util.Log("Processed submission", this.sub, LOG_PROCESSOR)
	doneChan <- E{}
}

func (this *Processor) ProcessFile(fileId bson.ObjectId) (err error) {
	//Retrieve file and process it.
	file, err := db.File(bson.M{db.ID: fileId}, nil)
	if err != nil {
		return
	}
	switch file.Type {
	case project.TEST:
		err = this.ProcessTest(file)
	case project.ARCHIVE:
		err = this.ProcessArchive(file)
	case project.SRC:
		err = this.ProcessSource(file)
	default:
		err = fmt.Errorf("Cannot process file type %s.", file.Type)
	}
	return
}

//ProcessFile extracts archives and runs tools on source files.
func (this *Processor) ProcessSource(file *project.File) (err error) {
	util.Log("Processing file:", file, LOG_PROCESSOR)
	defer util.Log("Processed file:", file, err, LOG_PROCESSOR)
	//Create a target for the tools to run on and save the file.
	target := tool.NewTarget(file.Name, file.Package, this.srcDir, tool.Language(this.project.Lang))
	err = util.SaveFile(target.FilePath(), file.Data)
	if err == nil {
		this.RunTools(file, target)
	}
	return
}

func (this *Processor) ProcessTest(test *project.File) (err error) {
	target := tool.NewTarget(test.Name, test.Package, this.srcDir, tool.Language(this.project.Lang))
	err = util.SaveFile(target.FilePath(), test.Data)
	if err != nil {
		return
	}
	junitJar, err := config.JUNIT.Path()
	if err != nil {
		return
	}
	this.compiler.(*javac.Tool).AddCP(junitJar)
	err = this.Compile(test.Id, target)
	if err != nil {
		return
	}
	testProc, err := NewTestProcessor(test, this)
	if err != nil {
		return
	}
	files, err := db.Files(bson.M{db.SUBID: test.SubId, db.TYPE: project.SRC}, bson.M{db.ID: 1})
	if err != nil {
		return
	}
	for _, f := range files {
		perr := testProc.Process(f.Id)
		if err != nil {
			util.Log(perr)
		}
	}
	return
}

//Extract extracts files from an archive, stores and processes them.
func (this *Processor) ProcessArchive(archive *project.File) error {
	//Extract and store the files.
	extracted, err := util.UnzipToMap(archive.Data)
	if err != nil {
		return err
	}
	for name, data := range extracted {
		err = this.StoreFile(name, data)
		if err != nil {
			util.Log(err, LOG_PROCESSOR)
		}
	}
	//We don't need the archive anymore
	err = db.RemoveFileById(archive.Id)
	if err != nil {
		util.Log(err, LOG_PROCESSOR)
	}
	files, err := db.Files(bson.M{db.SUBID: this.sub.Id}, bson.M{db.TIME: 1, db.ID: 1}, db.TIME)
	if err != nil {
		return err
	}
	ChangeStatus(Status{len(files), 0})
	//Process archive files.
	for _, file := range files {
		err = this.ProcessFile(file.Id)
		if err != nil {
			util.Log(err, LOG_PROCESSOR)
		}
		ChangeStatus(Status{-1, 0})
	}
	return nil
}

//StoreFile creates a new project.File given an encoded file name and file data.
//The new project.File is then saved in the database.
func (this *Processor) StoreFile(name string, data []byte) (err error) {
	file, err := project.ParseName(name)
	if err != nil {
		return
	}
	matcher := bson.M{db.SUBID: this.sub.Id, db.TYPE: file.Type, db.TIME: file.Time}
	if !db.Contains(db.FILES, matcher) {
		file.SubId = this.sub.Id
		file.Data = data
		err = db.Add(db.FILES, file)
	}
	return
}

//RunTools runs all available tools on a file. It skips a tool if
//there is already a result for it present. This makes it possible to
//rerun old tools or add new tools and run them on old files without having
//to rerun all the tools.
func (this *Processor) RunTools(file *project.File, target *tool.Target) {
	//First we compile our file.
	err := this.Compile(file.Id, target)
	if err != nil {
		util.Log(err, LOG_PROCESSOR)
		return
	}
	for _, t := range this.tools {
		//Skip the tool if it has already been run.
		if _, ok := file.Results[t.Name()]; ok {
			continue
		}
		res, err := t.Run(file.Id, target)
		if err != nil {
			//Report any errors and store timeouts.
			util.Log(
				fmt.Errorf("Encountered error %q when running tool %s on file %s.",
					err, t.Name(), file.Id.Hex()), LOG_PROCESSOR)
			if tool.IsTimeout(err) {
				err = db.AddFileResult(file.Id, t.Name(), tool.TIMEOUT)
			} else {
				err = db.AddFileResult(file.Id, t.Name(), tool.ERROR)
			}
		} else if res != nil {
			err = db.AddResult(res, t.Name())
		} else {
			err = db.AddFileResult(file.Id, t.Name(), tool.NORESULT)
		}
		if err != nil {
			util.Log(err, LOG_PROCESSOR)
		}
	}
}

//Compile compiles a file, stores the result and returns any errors which may have occured.
func (this *Processor) Compile(fileId bson.ObjectId, target *tool.Target) (err error) {
	res, err := this.compiler.Run(fileId, target)
	//We want to store the result if it is a compilation error
	if err != nil && !tool.IsCompileError(err) {
		return
	}
	db.AddResult(res, this.compiler.Name())
	return
}

func NewTestProcessor(testFile *project.File, parent *Processor) (proc *TestProcessor, err error) {
	dir := filepath.Join(parent.rootDir, testFile.Id.Hex())
	toolDir := filepath.Join(dir, "tools")
	compiler, err := javac.New("")
	if err != nil {
		return
	}
	runnerDir, err := config.JUNIT_TESTING.Path()
	if err != nil {
		return
	}
	err = util.Copy(toolDir, runnerDir)
	if err != nil {
		return
	}
	test := &junit.Test{
		Id:      testFile.Id,
		Name:    testFile.Name,
		Package: testFile.Package,
		Test:    testFile.Data,
	}
	testTool, err := junit.New(test, toolDir)
	if err != nil {
		return
	}
	testDir := filepath.Join(toolDir, test.Id.Hex())
	srcDir := filepath.Join(dir, "src")
	err = os.MkdirAll(srcDir, util.DPERM)
	if err != nil {
		return
	}
	testTarget := tool.NewTarget(test.Name, test.Package, testDir, tool.JAVA)
	cov, err := jacoco.New(dir, srcDir, testTarget)
	if err != nil {
		return
	}
	proc = &TestProcessor{
		id:       test.Id,
		results:  testFile.Results,
		rootDir:  dir,
		srcDir:   srcDir,
		toolDir:  toolDir,
		compiler: compiler,
		tools:    []tool.Tool{testTool, cov},
	}
	return
}

func (this *TestProcessor) Process(fileId bson.ObjectId) (err error) {
	file, err := db.File(bson.M{db.ID: fileId}, nil)
	if err != nil {
		return
	}
	target := tool.NewTarget(file.Name, file.Package, this.srcDir, tool.JAVA)
	err = util.SaveFile(target.FilePath(), file.Data)
	if err != nil {
		return
	}
	_, err = this.compiler.Run(file.Id, target)
	if err != nil {
		return
	}
	for _, t := range this.tools {
		rerr := this.Run(t, file, target)
		if err != nil {
			util.Log(rerr, LOG_PROCESSOR)
		}
	}
	return
}

func (this *TestProcessor) Run(t tool.Tool, file *project.File, target *tool.Target) (err error) {
	name := t.Name() + "-" + file.Id.Hex()
	if _, ok := this.results[name]; ok {
		return
	}
	res, rerr := t.Run(this.id, target)
	if rerr != nil {
		util.Log(rerr, LOG_PROCESSOR)
		//Report any errors and store timeouts.
		if tool.IsTimeout(rerr) {
			err = db.AddFileResult(this.id, name, tool.TIMEOUT)
		} else {
			err = db.AddFileResult(file.Id, name, tool.ERROR)
		}
	} else if res != nil {
		err = db.AddResult(res, name)
	} else {
		err = db.AddFileResult(this.id, name, tool.NORESULT)
	}
	return
}
