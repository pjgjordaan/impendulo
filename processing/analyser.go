package processing

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
)

const (
	LOG_ANALYSER = "processing/analyser.go"
)

//Analyser is used to run tools on a file.
type Analyser struct {
	proc   *Processor
	file   *project.File
	target *tool.TargetInfo
}

//Eval builds and runs tools on a source file.
func (this *Analyser) Eval() (err error) {
	err = this.buildTarget()
	if err != nil {
		return
	}
	this.RunTools()
	return
}

//buildTarget saves a file to filesystem.
func (this *Analyser) buildTarget() error {
	this.target = tool.NewTarget(this.file.Name,
		this.proc.project.Lang, this.file.Package, this.proc.srcDir)
	return util.SaveFile(this.target.FilePath(), this.file.Data)
}

//RunTools runs all available tools on a file, skipping previously run tools.
func (this *Analyser) RunTools() {
	//First we compile our file.
	res, err := this.proc.compiler.Run(this.file.Id, this.target)
	//If an error occurred we don't run any tools.
	if err != nil {
		//If the error was only compilation we still want to store the result.
		if tool.IsCompileError(err) {
			db.AddResult(res)
		}
		return
	}
	db.AddResult(res)
	for _, t := range this.proc.tools {
		//Skip the tool if it has already been run.
		if _, ok := this.file.Results[t.Name()]; ok {
			continue
		}
		res, err := t.Run(this.file.Id, this.target)
		if err != nil {
			//Report any errors and store timeouts.
			util.Log(
				fmt.Errorf("Encountered error %q when running tool %s on file %s.",
					err, t.Name(), this.file.Id.Hex()), LOG_ANALYSER)
			if tool.IsTimeout(err) {
				err = db.AddTimeoutResult(
					this.file.Id, t.Name())
			} else {
				continue
			}
		} else if res != nil {
			err = db.AddResult(res)
		} else {
			err = db.AddNoResult(this.file.Id, t.Name())
		}
		if err != nil {
			util.Log(err, LOG_ANALYSER)
		}
	}
}
