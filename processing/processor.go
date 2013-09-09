package processing

import (
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

//Processor is used to process individual submissions.
type Processor struct {
	sub      *project.Submission
	project  *project.Project
	rootDir  string
	srcDir   string
	toolDir  string
	jpfPath  string
	compiler tool.Tool
	tools    []tool.Tool
}

func NewProcessor(subId bson.ObjectId) (proc *Processor, err error) {
	sub, err := db.GetSubmission(bson.M{project.ID: subId}, nil)
	if err != nil {
		return
	}
	matcher := bson.M{project.ID: sub.ProjectId}
	proj, err := db.GetProject(matcher, nil)
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
	proc.compiler, err = Compiler(proc)
	if err != nil {
		return
	}
	proc.tools, err = Tools(proc)
	return
}

//Process processes a new submission.
//It listens for incoming files and processes them.
func (this *Processor) Process(fileChan chan bson.ObjectId, doneChan chan interface{}) {
	util.Log("Processing submission", this.sub)
	defer os.RemoveAll(this.rootDir)
	//Processing loop.
processing:
	for {
		select {
		case fId := <-fileChan:
			//Retrieve file and process it.
			file, err := db.GetFile(bson.M{project.ID: fId}, nil)
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

//ProcessFile processes a file according to its type.
func (this *Processor) ProcessFile(file *project.File) (err error) {
	util.Log("Processing file:", file)
	switch file.Type {
	case project.ARCHIVE:
		err = this.Extract(file)
	case project.SRC:
		analyser := &Analyser{
			proc: this,
			file: file,
		}
		err = analyser.Eval()
	}
	util.Log("Processed file:", file, err)
	return
}

//Extract extracts files from an archive and processes them.
func (this *Processor) Extract(archive *project.File) error {
	files, err := util.UnzipToMap(archive.Data)
	if err != nil {
		return err
	}
	for name, data := range files {
		err = this.storeFile(name, data)
		if err != nil {
			util.Log(err, LOG_PROCESSOR)
		}
	}
	err = db.RemoveFileById(archive.Id)
	if err != nil {
		util.Log(err, LOG_PROCESSOR)
	}
	fIds, err := db.GetFiles(bson.M{project.SUBID: this.sub.Id},
		bson.M{project.TIME: 1, project.ID: 1}, project.TIME)
	if err != nil {
		return err
	}
	ChangeStatus(Status{len(fIds), 0})
	//Process archive files.
	for _, fId := range fIds {
		file, err := db.GetFile(bson.M{project.ID: fId.Id}, nil)
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

func (this *Processor) storeFile(name string, data []byte) (err error) {
	file, err := project.ParseName(name)
	if err != nil {
		return
	}
	matcher := bson.M{project.SUBID: this.sub.Id, project.TYPE: file.Type, project.TIME: file.Time}
	if !db.Contains(db.FILES, matcher) {
		file.SubId = this.sub.Id
		file.Data = data
		err = db.AddFile(file)
	}
	return
}
