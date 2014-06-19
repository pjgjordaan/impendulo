package processor

import (
	"fmt"

	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"

	"testing"
	"time"
)

func TestFull(t *testing.T) {
	testFull(t, 1, 1, 1)
	testFull(t, 100, 1, 1)
	testFull(t, 10, 1, 2)
	testFull(t, 10, 2, 1)
	testFull(t, 10, 10, 10)
	testFull(t, 100, 10, 10)
}

func testFull(t *testing.T, nFiles, nProducers, nConsumers int) {
	var e error
	if e = MonitorStatus(); e != nil {
		t.Error(e)
	}
	rChan := make(chan *request.R)
	actualConsumers := 2 * nConsumers
	handlers := make([]*mq.MessageHandler, actualConsumers)
	for i := 0; i < nConsumers; i++ {
		if handlers[2*i+1], handlers[2*i], e = mq.NewSubmitter(rChan); e != nil {
			t.Error(e)
		}
	}
	for _, h := range handlers {
		go mq.H(h)
	}
	time.Sleep(2 * time.Second)
	for i := 0; i < nProducers; i++ {
		go func() {
			subId := bson.NewObjectId()
			key, e := mq.StartSubmission(subId)
			if e != nil {
				t.Error(e)
			}
			for i := 0; i < nFiles; i++ {
				file := &project.File{
					Id:    bson.NewObjectId(),
					SubId: subId,
					Type:  project.SRC,
				}
				if e = mq.AddFile(file, key); e != nil {
					t.Error(e)
				}
			}
			if e = mq.EndSubmission(subId, key); e != nil {
				t.Error(e)
			}
		}()
	}
	doneCount := 0
loop:
	for r := range rChan {
		if e = mq.ChangeStatus(r); e != nil {
			t.Error(e)
		}
		switch r.Type {
		case request.SUBMISSION_STOP:
			doneCount++
			if doneCount >= nProducers {
				break loop
			}
		case request.FILE_ADD:
			r.Type = request.FILE_REMOVE
			if e = mq.ChangeStatus(r); e != nil {
				t.Error(e)
			}
		}
	}
	if e = mq.WaitIdle(); e != nil {
		t.Error(e)
	}
	for _, h := range handlers {
		if e = h.Shutdown(); e != nil {
			t.Error(e)
		}
	}
	if e = ShutdownMonitor(); e != nil {
		t.Error(e)
	}
	if e = mq.StopProducers(); e != nil {
		t.Error(e)
	}
	fmt.Printf("Completed for %d files, %d producers and %d consumers.\n", nFiles, nProducers, nConsumers)
}
