package processing

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"os"
	"path/filepath"
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

//Eval evaluates a source file by attempting to run tools on it.
func (this *Analyser) Eval() (err error) {
	err = this.buildTarget()
	if err != nil {
		return
	}
	this.RunTools()
	return
}

//buildTarget saves a file to filesystem.
//It returns file info used by tools.
func (this *Analyser) buildTarget() error {
	this.target = tool.NewTarget(this.file.Name,
		this.proc.project.Lang, this.file.Package, this.proc.srcDir)
	return util.SaveFile(this.target.FilePath(), this.file.Data)
}

//RunTools runs all available tools on a file, skipping previously run tools.
func (this *Analyser) RunTools() {
	res, err := this.proc.compiler.Run(this.file.Id, this.target)
	if err != nil {
		if tool.IsCompileError(err) {
			db.AddResult(res)
		}
		return
	}
	db.AddResult(res)
	for _, t := range this.proc.tools {
		if _, ok := this.file.Results[t.Name()]; ok {
			continue
		}
		res, err := t.Run(this.file.Id, this.target)
		if err != nil {
			util.Log(
				fmt.Errorf("Encountered error %q when running tool %s on file %s.",
					err, t.Name(), this.file.Id.Hex()), LOG_ANALYSER)
			dest := filepath.Join(os.TempDir(), "errors", this.file.Id.Hex())
			os.MkdirAll(dest, util.DPERM)
			util.Copy(dest, this.target.Dir)
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
