package processing

import (
	"container/list"
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

const LOG_SERVER = "processing/server.go"

var (
	fileChan, startSubmissionChan, endSubmissionChan chan *Request
	processedChan                                    chan interface{}
	statusChan                                       chan Status
)

type Request struct {
	Id, ParentId bson.ObjectId
	Response     chan error
}

func init() {
	fileChan = make(chan *Request)
	startSubmissionChan = make(chan *Request)
	endSubmissionChan = make(chan *Request)
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

//AddFile sends a file id to be processed.
func AddFile(file *project.File) (err error) {
	if !file.CanProcess() {
		return
	}
	errChan := make(chan error)
	fileChan <- &Request{
		Id:       file.Id,
		ParentId: file.SubId,
		Response: errChan,
	}
	err = <-errChan
	if err != nil {
		return
	}
	ChangeStatus(Status{1, 0})
	return
}

//StartSubmission signals that this submission will now receive files.
func StartSubmission(subId bson.ObjectId) (err error) {
	errChan := make(chan error)
	startSubmissionChan <- &Request{
		Id:       subId,
		Response: errChan,
	}
	err = <-errChan
	if err != nil {
		return
	}
	ChangeStatus(Status{0, 1})
	return
}

//EndSubmission signals that this submission has stopped receiving files.
func EndSubmission(subId bson.ObjectId) error {
	errChan := make(chan error)
	endSubmissionChan <- &Request{
		Id:       subId,
		Response: errChan,
	}
	return <-errChan
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
			helper.SetStarted()
			if helper.Done() {
				delete(helpers, subId)
			}
			go helper.Handle(fileQueues[subId])
			delete(fileQueues, subId)
			busy++
		} else if busy < 0 {
			break
		}
		select {
		case request := <-fileChan:
			var err error
			if helper, ok := helpers[request.ParentId]; ok {
				if helper.started {
					//Send file to goroutine if
					//submission processing has started.
					helper.serveChan <- request.Id
				} else {
					//Add file to queue if not.
					fileQueues[request.ParentId].PushBack(request.Id)
				}
			} else {
				err = fmt.Errorf("No submission %q found for file %q.", request.ParentId, request.Id)
			}
			request.Response <- err
		case request := <-startSubmissionChan:
			var err error
			if _, ok := helpers[request.Id]; !ok {
				//Add submission to queue.
				subQueue.PushBack(request.Id)
				helpers[request.Id] = NewProcHelper(request.Id)
				fileQueues[request.Id] = list.New()
			} else {
				err = fmt.Errorf("Can't start submission %q, already being processed.", request.Id)
			}
			request.Response <- err
		case request := <-endSubmissionChan:
			var err error
			if helper, ok := helpers[request.Id]; ok {
				//Submission will receive no more files.
				helper.SetDone()
				if helper.Started() {
					delete(helpers, request.Id)
				}
			} else {
				err = fmt.Errorf("No submission %q found to end.", request.Id)
			}
			request.Response <- err
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

func (this *ProcHelper) Done() bool {
	return this.done
}

func (this *ProcHelper) SetStarted() {
	this.started = true
}

func (this *ProcHelper) Started() bool {
	return this.started
}

//Handle helps manage the files a submission receives.
//fileQueue is the queue of files the submission has received
//prior to the start of processing.
func (this *ProcHelper) Handle(fileQueue *list.List) {
	procChan := make(chan bson.ObjectId)
	stopChan := make(chan interface{})
	proc, err := NewProcessor(this.subId)
	if err != nil {
		util.Log(err, LOG_SERVER)
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
