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
	sub, err := db.Submission(bson.M{project.ID: subId}, nil)
	if err != nil {
		return
	}
	matcher := bson.M{project.ID: sub.ProjectId}
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
func (this *Processor) Process(fileChan chan bson.ObjectId, doneChan chan interface{}) {
	util.Log("Processing submission", this.sub)
	defer os.RemoveAll(this.rootDir)
	//Processing loop.
processing:
	for {
		select {
		case fId := <-fileChan:
			//Retrieve file and process it.
			file, err := db.File(bson.M{project.ID: fId}, nil)
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
	util.Log("Processed submission", this.sub)
}

//ProcessFile extracts archives and runs tools on source files.
func (this *Processor) ProcessFile(file *project.File) (err error) {
	util.Log("Processing file:", file)
	defer util.Log("Processed file:", file, err)
	switch file.Type {
	case project.ARCHIVE:
		err = this.Extract(file)
	case project.SRC:
		//Create a target for the tools to run on and save the file.
		target := tool.NewTarget(file.Name,
			this.project.Lang, file.Package, this.srcDir)
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
	fIds, err := db.Files(bson.M{project.SUBID: this.sub.Id},
		bson.M{project.TIME: 1, project.ID: 1}, project.TIME)
	if err != nil {
		return err
	}
	ChangeStatus(Status{len(fIds), 0})
	//Process archive files.
	for _, fId := range fIds {
		file, err := db.File(bson.M{project.ID: fId.Id}, nil)
		if err != nil {
			util.Log(err, LOG_PROCESSOR)
		} else {
			err = this.ProcessFile(file)
			if err != nil {
				util.Log(err, LOG_PROCESSOR)
			}
		}
		fileProcessed()
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
	matcher := bson.M{project.SUBID: this.sub.Id, project.TYPE: file.Type, project.TIME: file.Time}
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
func (this *Processor) RunTools(file *project.File, target *tool.TargetInfo) {
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
				err = db.AddFileResult(
					file.Id, t.Name(), tool.TIMEOUT)
			} else {
				err = db.AddFileResult(
					file.Id, t.Name(), tool.ERROR)
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
func (this *Processor) Compile(fileId bson.ObjectId, target *tool.TargetInfo) (err error) {
	res, err := this.compiler.Run(fileId, target)
	//We want to store the result if it is a compilation error
	if err != nil && !tool.IsCompileError(err) {
		return
	}
	db.AddResult(res)
	return
}