package processor

import (
	"container/list"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/worker/file"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	//Handler is used to help handle a submission's files.
	Handler struct {
		subId                bson.ObjectId
		testChan, fileChan   chan bson.ObjectId
		fileQueue, testQueue *list.List
		doneChan             chan util.E
		started              bool
		done                 bool
	}
)

//NewHandler creates a new Handler for the specified
//Submission.
func NewHandler(sid bson.ObjectId) *Handler {
	return &Handler{
		subId:     sid,
		fileChan:  make(chan bson.ObjectId),
		testChan:  make(chan bson.ObjectId),
		fileQueue: list.New(),
		testQueue: list.New(),
		doneChan:  make(chan util.E),
		started:   false,
		done:      false,
	}
}

func (h *Handler) AddFile(r *request.R) {
	switch r.Type {
	case request.SRC_ADD, request.ARCHIVE_ADD:
		if h.started {
			h.fileChan <- r.FileId
		} else {
			h.fileQueue.PushBack(r.FileId)
		}
	case request.TEST_ADD:
		if h.started {
			h.testChan <- r.FileId
		} else {
			h.testQueue.PushBack(r.FileId)
		}
	}
}

//SetDone indicates that this submission will receive no more files.
func (h *Handler) SetDone() {
	if h.started {
		//If it has started send on its channel.
		h.doneChan <- util.E{}
	} else {
		//Otherwise simply set done to true.
		h.done = true
	}
}

//Handle helps manage the files a submission receives.
//It spawns a new Processor which runs in a seperate goroutine
//and receives files to process from this Handler.
//fq is the queue of files the submission has received
//prior to the start of processing.
func (h *Handler) Handle(onDone chan util.E) {
	defer func() {
		if e := mq.ChangeStatus(request.StopSubmission(h.subId)); e != nil {
			util.Log(e, LOG_SERVER)
		}
		onDone <- util.E{}
	}()
	w, e := file.New(h.subId)
	if e != nil {
		util.Log(e, LOG_SERVER)
		return
	}
	pc := make(chan bson.ObjectId)
	sc := make(chan util.E)
	go w.Start(pc, sc)
	busy := false
	for {
		if !busy {
			if fid := h.nextFile(); fid != "" {
				pc <- fid
				busy = true
			} else if h.done {
				sc <- util.E{}
				return
			}
		}
		select {
		case fid := <-h.fileChan:
			//Add new files to the queue.
			h.fileQueue.PushBack(fid)
		case fid := <-h.testChan:
			//Add new files to the queue.
			h.testQueue.PushBack(fid)
		case fid := <-pc:
			if e := removeFile(fid); e != nil {
				util.Log(e)
			}
			busy = false
		case <-h.doneChan:
			//Submission will receive no more files.
			h.done = true
		}
	}
}

func (h *Handler) nextFile() bson.ObjectId {
	if h.fileQueue.Len() > 0 {
		return h.fileQueue.Remove(h.fileQueue.Front()).(bson.ObjectId)
	} else if h.done && h.testQueue.Len() > 0 {
		return h.testQueue.Remove(h.testQueue.Front()).(bson.ObjectId)
	}
	return ""
}

func removeFile(fid bson.ObjectId) error {
	f, e := db.File(bson.M{db.ID: fid}, db.FILE_SELECTOR)
	if e != nil {
		return e
	}
	r, e := request.RemoveFile(f)
	if e != nil {
		return e
	}
	if e = mq.ChangeStatus(r); e != nil {
		return e
	}
	if r.Type != request.ARCHIVE_REMOVE {
		return nil
	}
	return db.RemoveFileById(fid)
}
