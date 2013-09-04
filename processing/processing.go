package processing

import (
	"container/list"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

const LOG_PROCESSING = "processing/processing.go"

var idChan chan *ids
var processedChan chan interface{}

var statusChan chan Status

func init() {
	idChan = make(chan *ids)
	processedChan = make(chan interface{})
	statusChan = make(chan Status)
}

//Status is used to indicate a change in the files or
//submissions being processed. It is also used to retrieve the current
//number of files and submissions being processed.
type Status struct {
	Files       int
	Submissions int
}

func (this *Status) add(toAdd Status) {
	this.Files += toAdd.Files
	this.Submissions += toAdd.Submissions
}

//ChangeStatus is used to update Impendulo's current
//processing status.
func ChangeStatus(change Status) {
	statusChan <- change
}

//GetStatus retrieves Impendulo's current processing status.
func GetStatus() (ret Status) {
	statusChan <- Status{0, 0}
	ret = <-statusChan
	return
}

func monitorStatus() {
	var status *Status = new(Status)
	for {
		val := <-statusChan
		switch val {
		case Status{0, 0}:
			statusChan <- *status
		default:
			status.add(val)
		}
	}
}

type ids struct {
	fileId bson.ObjectId
	subId  bson.ObjectId
	isFile bool
}

//AddFile sends a file id to be processed.
func AddFile(file *project.File) {
	if file.CanProcess() {
		idChan <- &ids{file.Id, file.SubId, true}
		ChangeStatus(Status{1, 0})
	}
}

//StartSubmission signals that this submission will now receive files.
func StartSubmission(subId bson.ObjectId) {
	idChan <- &ids{subId: subId, isFile: false}
	ChangeStatus(Status{0, 1})
}

//EndSubmission signals that this submission has stopped receiving files.
func EndSubmission(subId bson.ObjectId) {
	idChan <- &ids{subId: subId, isFile: false}
}

func submissionProcessed() {
	ChangeStatus(Status{0, -1})
	processedChan <- None()
}

func fileProcessed() {
	ChangeStatus(Status{-1, 0})
}

//Shutdown stops Serve from running once all submissions have been processed.
func Shutdown() {
	processedChan <- None()
}

//None provides an empty struct
func None() interface{} {
	type e struct{}
	return e{}
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
func Serve(maxProcs int) {
	go monitorStatus()
	helpers := make(map[bson.ObjectId]*ProcHelper)
	fileQueues := make(map[bson.ObjectId]*list.List)
	subQueue := list.New()
	busy := 0
	for {
		if busy < maxProcs && subQueue.Len() > 0 {
			//If there is an available spot,
			//start processing the next submission.
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
						//Send file to goroutine if
						//submission processing has started.
						helper.serveChan <- ids.fileId
					} else {
						//Add file to queue if not.
						fileQueues[ids.subId].PushBack(ids.fileId)
					}
				} else {
					util.Log(fmt.Errorf(
						"No submission %q found for file %q.",
						ids.subId, ids.fileId), LOG_PROCESSING)
				}
			} else {
				if helper, ok := helpers[ids.subId]; ok {
					//Submission will receive no more files.
					helper.SetDone()
				} else {
					//Add submission to queue.
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
	return &ProcHelper{subId, make(chan bson.ObjectId),
		make(chan interface{}), false, false}
}

//ProcHelper is used to help handle a submission's files.
type ProcHelper struct {
	subId     bson.ObjectId
	serveChan chan bson.ObjectId
	doneChan  chan interface{}
	started   bool
	done      bool
}

//SetDone indicates that this submission will receive no more files.
func (this *ProcHelper) SetDone() {
	if this.started {
		this.doneChan <- None()
	} else {
		this.done = true
	}
}

//Handle helps manage the files a submission receives.
//fileQueue is the queue of files the submission has received
//prior to the start of processing.
func (this *ProcHelper) Handle(fileQueue *list.List) {
	procChan := make(chan bson.ObjectId)
	stopChan := make(chan interface{})
	proc, err := NewProcessor(this.subId)
	if err != nil {
		util.Log(err, LOG_PROCESSING)
		submissionProcessed()
		return
	}
	go proc.Process(procChan, stopChan)
	busy := false
	for {
		if !busy {
			if fileQueue.Len() > 0 {
				//Not busy so send a new File to be processed.
				fId := fileQueue.Remove(
					fileQueue.Front()).(bson.ObjectId)
				procChan <- fId
				busy = true
			} else if this.done {
				//Not busy and done so we can finish up here.
				stopChan <- None()
				submissionProcessed()
				return
			}
		}
		select {
		case fId := <-this.serveChan:
			//Add new files to the queue.
			fileQueue.PushBack(fId)
		case <-procChan:
			//Processor has finished with its current file.
			busy = false
			fileProcessed()
		case <-this.doneChan:
			//Submission will receive no more files.
			this.done = true
		}
	}
}

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
				util.Log(err, LOG_PROCESSING)
			} else {
				err = this.ProcessFile(file)
				if err != nil {
					util.Log(err, LOG_PROCESSING)
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
			util.Log(err, LOG_PROCESSING)
		}
	}
	err = db.RemoveFileById(archive.Id)
	if err != nil {
		util.Log(err, LOG_PROCESSING)
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
			util.Log(err, LOG_PROCESSING)
		} else {
			err = this.ProcessFile(file)
			if err != nil {
				util.Log(err, LOG_PROCESSING)
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
			util.Log(err, LOG_PROCESSING)
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
			util.Log(err, LOG_PROCESSING)
		}
	}
}
