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
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

const (
	LOG_PROCESSOR = "processing/processor.go"
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
	//defer os.RemoveAll(this.rootDir)
	//Processing loop.
processing:
	for {
		select {
		case fId := <-fileChan:
			//Retrieve file and process it.
			file, err := db.File(bson.M{db.ID: fId}, nil)
			if err != nil {
				util.Log(err, LOG_PROCESSOR)
			} else {
				err = this.ProcessFile(file)
				if err != nil {
					util.Log(err, LOG_PROCESSOR)
				}
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

//ProcessFile extracts archives and runs tools on source files.
func (this *Processor) ProcessFile(file *project.File) (err error) {
	util.Log("Processing file:", file, LOG_PROCESSOR)
	defer util.Log("Processed file:", file, err, LOG_PROCESSOR)
	switch file.Type {
	case project.ARCHIVE:
		err = this.Extract(file)
	case project.SRC:
		//Create a target for the tools to run on and save the file.
		target := tool.NewTarget(file.Name, file.Package, this.srcDir, tool.Language(this.project.Lang))
		err = util.SaveFile(target.FilePath(), file.Data)
		if err == nil {
			this.RunTools(file, target)
		}
	}
	return
}

//Extract extracts files from an archive, stores and processes them.
func (this *Processor) Extract(archive *project.File) error {
	//Extract and store the files.
	files, err := util.UnzipToMap(archive.Data)
	if err != nil {
		return err
	}
	for name, data := range files {
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
	fIds, err := db.Files(bson.M{db.SUBID: this.sub.Id},
		bson.M{db.TIME: 1, db.ID: 1}, db.TIME)
	if err != nil {
		return err
	}
	ChangeStatus(Status{len(fIds), 0})
	//Process archive files.
	for _, fId := range fIds {
		file, err := db.File(bson.M{db.ID: fId.Id}, nil)
		if err != nil {
			util.Log(err, LOG_PROCESSOR)
		} else {
			err = this.ProcessFile(file)
			if err != nil {
				util.Log(err, LOG_PROCESSOR)
			}
		}
		//fileProcessed()
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
			err = db.AddResult(res)
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
	db.AddResult(res)
	return
}
