package processing

import (
	"container/list"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

var idChan chan *ids
var processedChan chan interface{}

func init() {
	idChan = make(chan *ids)
	processedChan = make(chan interface{})
}

type ids struct {
	fileId bson.ObjectId
	subId  bson.ObjectId
	isFile bool
}

//AddFile sends a file id to be processed.
func AddFile(file *project.File) {
	if file.CanProcess(){
		idChan <- &ids{file.Id, file.SubId, true}
	}
}

//EndSubmission stops this submission's goroutine to.
func StartSubmission(subId bson.ObjectId) {
	idChan <- &ids{subId: subId, isFile: false}
}

func EndSubmission(subId bson.ObjectId) {
	idChan <- &ids{subId: subId, isFile: false}
}

func submissionProcessed() {
	processedChan <- None()
}

func Shutdown() {
	processedChan <- None()
}

func None() interface{} {
	type e struct{}
	return e{}
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
//Incomplete submissions are read from disk and reprocessed via ProcessStored.
func Serve(maxProcs int) {
	helpers := make(map[bson.ObjectId]*ProcHelper)
	fileQueues := make(map[bson.ObjectId]*list.List)
	subQueue := list.New()
	busy := 0
	for {
		if busy < maxProcs && subQueue.Len() > 0 {
			subId := subQueue.Remove(subQueue.Front()).(bson.ObjectId)
			helper := helpers[subId]
			helper.started = true
			go helper.Handle(fileQueues[subId])
			delete(fileQueues, subId)
			busy++
		} else if busy < 0 {
			break
		}
		select {
		case ids := <-idChan:
			if ids.isFile {
				if helper, ok := helpers[ids.subId]; ok {
					if helper.started {
						helper.serveChan <- ids.fileId
					} else {
						fileQueues[ids.subId].PushBack(ids.fileId)
					}
				} else {
					util.Log(fmt.Errorf("No submission %q found for file %q.", ids.subId, ids.fileId))
				}
			} else {
				if helper, ok := helpers[ids.subId]; ok {
					helper.SetDone()
				} else {
					subQueue.PushBack(ids.subId)
					helpers[ids.subId] = NewProcHelper(ids.subId)
					fileQueues[ids.subId] = list.New()
				}
			}
		case <-processedChan:
			busy--
		}
	}
}

func NewProcHelper(subId bson.ObjectId) *ProcHelper {
	return &ProcHelper{subId, make(chan bson.ObjectId), make(chan interface{}), false, false}
}

type ProcHelper struct {
	subId     bson.ObjectId
	serveChan chan bson.ObjectId
	doneChan  chan interface{}
	started   bool
	done      bool
}

func (this *ProcHelper) SetDone() {
	if this.started {
		this.doneChan <- None()
	} else {
		this.done = true
	}
}

func (this *ProcHelper) Handle(fileQueue *list.List) {
	procChan := make(chan bson.ObjectId)
	stopChan := make(chan interface{})
	proc, err := NewProcessor(this.subId)
	if err != nil {
		util.Log(err)
	}
	go proc.Process(procChan, stopChan)
	busy := false
	for {
		if !busy {
			if fileQueue.Len() > 0 {
				fId := fileQueue.Remove(fileQueue.Front()).(bson.ObjectId)
				procChan <- fId
				busy = true
			} else if this.done {
				stopChan <- None()
				submissionProcessed()
				return
			}
		}
		select {
		case fId := <-this.serveChan:
			fileQueue.PushBack(fId)
		case <-procChan:
			busy = false
		case <-this.doneChan:
			this.done = true
		}
	}
}

//Processor is used to process individual submissions.
type Processor struct {
	sub     *project.Submission
	tests   []*TestRunner
	rootDir string
	srcDir  string
	toolDir string
	jpfPath string
}

func NewProcessor(subId bson.ObjectId) (proc *Processor, err error) {
	sub, err := db.GetSubmission(bson.M{project.ID: subId}, nil)
	if err != nil {
		return
	}
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	proc = &Processor{sub: sub, rootDir: dir, srcDir: filepath.Join(dir, "src"), toolDir: filepath.Join(dir, "tools")}
	return
}

//Process processes a new submission.
//It listens for incoming files and creates new goroutines to processes them.
func (this *Processor) Process(fileChan chan bson.ObjectId, doneChan chan interface{}) {
	util.Log("Processing submission", this.sub)
	defer os.RemoveAll(this.rootDir)
	err := this.SetupJPF()
	if err != nil {
		util.Log(err)
	}
	this.tests, err = SetupTests(this.sub.ProjectId, this.toolDir)
	if err != nil {
		util.Log(err)
	}
outer:
	for {
		select {
		case fId := <-fileChan:
			file, err := db.GetFile(bson.M{project.ID: fId}, nil)
			if err != nil {
				util.Log(err)
			} else {
				err = this.ProcessFile(file)
				if err != nil {
					util.Log(err)
				}
			}
			fileChan <- fId
		case <-doneChan:
			break outer
		}
	}
	util.Log("Processed submission", this.sub)
}

//Setup sets up the environment needed for this Processor to function correctly.
func (this *Processor) SetupJPF() error {
	err := util.Copy(this.toolDir, config.GetConfig(config.RUNNER_DIR))
	if err != nil {
		return err
	}
	jpfFile, err := db.GetJPF(bson.M{project.PROJECT_ID: this.sub.ProjectId}, nil)
	if err != nil {
		return err
	}
	this.jpfPath = filepath.Join(this.toolDir, jpfFile.Name)
	return util.SaveFile(this.jpfPath, jpfFile.Data)
}

//ProcessFile processes a file according to its type.
func (this *Processor) ProcessFile(file *project.File) (err error) {
	util.Log("Processing file:", file)
	switch file.Type {
	case project.ARCHIVE:
		err = this.extract(file)
	case project.SRC:
		analyser := &Analyser{proc: this, file: file}
		err = analyser.Eval()
	}
	util.Log("Processed file:", file, err)
	return
}

//ProcessArchive extracts files from an archive and processes them.
func (this *Processor) extract(archive *project.File) error {
	files, err := util.UnzipToMap(archive.Data)
	if err != nil {
		return err
	}
	for name, data := range files {
		err = storeFile(name, data, this.sub.Id) 
		if err != nil {
			util.Log(err)
		}
	}
	err = db.RemoveFileById(archive.Id)
	if err != nil {
		util.Log(err)
	}
	fIds, err := db.GetFiles(bson.M{project.SUBID: this.sub.Id}, 
		bson.M{project.NUM: 1, project.ID: 1}, project.NUM)
	if err != nil {
		return err
	}
	for _, fId := range fIds {
		file, err := db.GetFile(bson.M{project.ID: fId.Id}, nil)
		if err != nil {
			util.Log(err)
			continue
		}
		err = this.ProcessFile(file)
		if err != nil {
			util.Log(err)
		}
	}
	return nil
}

func storeFile(name string, data []byte, subId bson.ObjectId)(err error){
	file, err := project.ParseName(name)
	if err != nil {
		return
	}
	matcher := bson.M{project.SUBID: subId, project.TIME: file.Time}
	if !db.Contains(db.FILES, matcher) {
		file.SubId = subId
		file.Data = data
		err = db.AddFile(file)
	}
	return
}

//Analyser is used to run tools on a file.
type Analyser struct {
	proc   *Processor
	file   *project.File
	target *tool.TargetInfo
}

//Eval evaluates a source or compiled file by attempting to run tests and tools on it.
func (this *Analyser) Eval() error {
	err := this.buildTarget()
	if err != nil {
		return err
	}
	var compileErr bool
	compileErr, err = this.compile()
	if err != nil {
		return err
	} else if compileErr {
		return nil
	}
	this.file, err = db.GetFile(bson.M{project.ID: this.file.Id}, nil)
	if err != nil {
		return err
	}
	for _, test := range this.proc.tests {
		err = test.Run(this.file, this.proc.srcDir)
		if err != nil {
			util.Log(err)
		}
	}
	this.RunTools()
	return nil
}

//buildTarget saves a file to filesystem.
//It returns file info used by tools & tests.
func (this *Analyser) buildTarget() error {
	matcher := bson.M{project.ID: this.proc.sub.ProjectId}
	p, err := db.GetProject(matcher, nil)
	if err != nil {
		return err
	}
	this.target = tool.NewTarget(this.file.Name, p.Lang, this.file.Package, this.proc.srcDir)
	return util.SaveFile(this.target.FilePath(), this.file.Data)
}

//compile compiles a java source file and saves the results thereof.
//It returns true if compiled successfully.
func (this *Analyser) compile() (bool, error) {
	comp := javac.NewJavac(this.target.Dir)
	res, err := comp.Run(this.file.Id, this.target)
	compileErr := javac.IsCompileError(err)
	if err != nil && !compileErr {
		return false, err
	}
	return compileErr, db.AddResult(res)
}

//RunTools runs all available tools on a file, skipping previously run tools.
func (this *Analyser) RunTools() {
	tools := []tool.Tool{findbugs.NewFindBugs(), pmd.NewPMD(),
		jpf.NewJPF(this.proc.toolDir, this.proc.jpfPath),
		checkstyle.NewCheckstyle()}
	for _, tool := range tools {
		if _, ok := this.file.Results[tool.GetName()]; ok {
			continue
		}
		res, err := tool.Run(this.file.Id, this.target)
		if err != nil {
			util.Log(err)
			continue
		}
		if res != nil {
			err = db.AddResult(res)
		}
		if err != nil {
			util.Log(err)
		}
	}
}
