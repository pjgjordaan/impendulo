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
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"github.com/streadway/amqp"
	"labix.org/v2/mgo/bson"
	"strconv"
	"testing"
	"time"
)

type (
	BasicConsumer struct {
		id   int
		msgs chan string
	}
)

func init() {
	fmt.Print()
	util.SetErrorLogging("f")
	util.SetInfoLogging("f")
}

func (bc *BasicConsumer) Consume(d amqp.Delivery, ch amqp.Channel) error {
	bc.msgs <- fmt.Sprintf("Consumer %d says %s.\n", bc.id, string(d.Body))
	d.Ack(false)
	return nil
}

func TestWaitIdle(t *testing.T) {
	statusChan := make(chan Status)
	w, err := NewWaiter(AMQP_URI, statusChan)
	if err != nil {
		t.Error(err)
	}
	go func() {
		err = w.Handle()
		if err != nil {
			t.Error(err)
		}
	}()
	status := Status{5, 5}
	go func() {
		for status.Files >= 0 {
			<-statusChan
			statusChan <- status
			status.add(Status{-1, -1})
		}
	}()
	err = WaitIdle()
	if err != nil {
		t.Error(err)
	}
	err = w.Shutdown()
	if err != nil {
		t.Error(err)
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestGetStatus(t *testing.T) {
	statusChan := make(chan Status)
	sl, err := NewLoader(AMQP_URI, statusChan)
	if err != nil {
		t.Error(err)
	}
	go func() {
		err = sl.Handle()
		if err != nil {
			t.Error(err)
		}
	}()
	status := Status{1, 1}
	go func() {
		<-statusChan
		statusChan <- status
	}()
	s, err := GetStatus()
	if err != nil {
		t.Error(err)
	} else if s.Files != status.Files || s.Submissions != status.Submissions {
		t.Error(fmt.Errorf("Invalid status %q.", s))
	}
	err = sl.Shutdown()
	if err != nil {
		t.Error(err)
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestFileConsume(t *testing.T) {
	subId := bson.NewObjectId()
	n := 10
	files := make([]*project.File, n)
	for i := 0; i < n; i++ {
		files[i] = &project.File{
			Id:    bson.NewObjectId(),
			SubId: subId,
			Type:  project.SRC,
		}
	}
	requestChan := make(chan *Request)
	handlers := make([]*MessageHandler, n)
	var err error
	for i := 0; i < n; i++ {
		handlers[i], err = NewFileConsumer(AMQP_URI, requestChan)
		if err != nil {
			t.Error(err)
		}
	}
	for _, h := range handlers {
		go func(fc *MessageHandler) {
			err = fc.Handle()
			if err != nil {
				t.Error(err)
			}
		}(h)
	}
	go func() {
		for _, file := range files {
			AddFile(file)
		}
	}()
	processed := 0
	for r := range requestChan {
		found := false
		for _, file := range files {
			if r.FileId == file.Id && r.SubId == file.SubId {
				found = true
				break
			}
		}
		if !found {
			t.Error(fmt.Errorf("Invalid request %q.", r))
		}
		processed++
		if processed == n {
			break
		}
	}
	for _, h := range handlers {
		err = h.Shutdown()
		if err != nil {
			t.Error(err)
		}

	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestSubmissionEnder(t *testing.T) {
	subId := bson.NewObjectId()
	n := 10
	requestChan := make(chan *Request)
	handlers := make([]*MessageHandler, n)
	var err error
	for i := 0; i < n; i++ {
		handlers[i], err = NewEnder(AMQP_URI, requestChan)
		if err != nil {
			t.Error(err)
		}
	}
	for _, h := range handlers {
		go func(fc *MessageHandler) {
			err = fc.Handle()
			if err != nil {
				t.Error(err)
			}
		}(h)
	}
	err = EndSubmission(subId)
	if err != nil {
		t.Error(err)
	}
	processed := 0
	for r := range requestChan {
		if r.SubId != subId || !r.Stop {
			t.Error(fmt.Errorf("Invalid request %q.", r))
		}
		processed++
		if processed == n {
			break
		}
	}
	for _, h := range handlers {
		err = h.Shutdown()
		if err != nil {
			t.Error(err)
		}
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestStatusChange(t *testing.T) {
	n := 10
	statusChan := make(chan Status)
	handlers := make([]*MessageHandler, n)
	var err error
	for i := 0; i < n; i++ {
		handlers[i], err = NewChanger(AMQP_URI, statusChan)
		if err != nil {
			t.Error(err)
		}
	}
	for _, h := range handlers {
		go func(fc *MessageHandler) {
			err = fc.Handle()
			if err != nil {
				t.Error(err)
			}
		}(h)
	}
	status := Status{1, 1}
	err = ChangeStatus(status)
	if err != nil {
		t.Error(err)
	}
	processed := 0
	for s := range statusChan {
		if s.Files != status.Files || s.Submissions != status.Submissions {
			t.Error(fmt.Errorf("Invalid status %q.", s))
		}
		processed++
		if processed == n {
			break
		}
	}
	for _, h := range handlers {
		err = h.Shutdown()
		if err != nil {
			t.Error(err)
		}
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestMonitorStatus(t *testing.T) {
	err := MonitorStatus(AMQP_URI)
	if err != nil {
		t.Error(err)
	}
	n := 10
	subId := bson.NewObjectId()
	files := make([]*project.File, n)
	for i := 0; i < n; i++ {
		files[i] = &project.File{
			Id:    bson.NewObjectId(),
			SubId: subId,
			Type:  project.SRC,
		}
	}
	requestChan := make(chan *Request)
	handlers := make([]*MessageHandler, n)
	for i := 0; i < n-1; i++ {
		handlers[i], err = NewFileConsumer(AMQP_URI, requestChan)
		if err != nil {
			t.Error(err)
		}
	}
	handlers[n-1], err = NewEnder(AMQP_URI, requestChan)
	if err != nil {
		t.Error(err)
	}
	for _, h := range handlers {
		go func(mh *MessageHandler) {
			err = mh.Handle()
			if err != nil {
				t.Error(err)
			}
		}(h)
	}
	go func() {
		for _, file := range files {
			err = AddFile(file)
			if err != nil {
				t.Error(err)
			}
		}
		err = EndSubmission(subId)
		if err != nil {
			t.Error(err)
		}
	}()
	stop := false
	submap := make(map[bson.ObjectId]struct{})
	fileCount := 0
	for r := range requestChan {
		if r.Stop {
			stop = true
		} else {
			if _, ok := submap[r.SubId]; ok {
				err = ChangeStatus(Status{1, 0})
			} else {
				submap[r.SubId] = struct{}{}
				err = ChangeStatus(Status{1, 1})
			}
			if err != nil {
				t.Error(err)
			}
			fileCount++
		}
		if fileCount == n && stop {
			break
		}
	}
	time.Sleep(1 * time.Second)
	err = ChangeStatus(Status{-fileCount, -len(submap)})
	if err != nil {
		t.Error(err)
	}
	err = WaitIdle()
	if err != nil {
		t.Error(err)
	}
	for _, h := range handlers {
		err = h.Shutdown()
		if err != nil {
			t.Error(err)
		}
	}
	err = ShutdownMonitor()
	if err != nil {
		t.Error(err)
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestAMQPBasic(t *testing.T) {
	msgChan := make(chan string)
	handler, err := NewHandler(AMQP_URI, "test", DIRECT, "", "", "", &BasicConsumer{id: 1, msgs: msgChan})
	if err != nil {
		t.Error(err)
	}
	producer, err := NewProducer("test_producer", AMQP_URI, "test", DIRECT, "")
	if err != nil {
		t.Error(err)
	}
	go func() {
		herr := handler.Handle()
		if herr != nil {
			t.Error(herr)
		}
	}()
	n := 10
	for i := 0; i < n; i++ {
		producer.Produce([]byte(fmt.Sprintf("testing %d", i)))
	}
	for i := 0; i < n; i++ {
		fmt.Printf("Message %d %s", i, <-msgChan)
	}
	err = handler.Shutdown()
	if err != nil {
		t.Error(err)
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}

func TestAMQPQueue(t *testing.T) {
	nP, nH, nM := 10, 10, 50
	msgChan := make(chan string)
	handlers := make([]*MessageHandler, nH)
	var err error
	for i := 0; i < nH; i++ {
		handlers[i], err = NewHandler(AMQP_URI, "test", DIRECT, "test_queue", "test_key", "", &BasicConsumer{id: i, msgs: msgChan})
		if err != nil {
			t.Error(err)
		}
	}
	producers := make([]*Producer, nP)
	for i := 0; i < nP; i++ {
		producers[i], err = NewProducer("test_producer_"+strconv.Itoa(i), AMQP_URI, "test", DIRECT, "test_key")
		if err != nil {
			t.Error(err)
		}
	}
	for _, handler := range handlers {
		go func(h *MessageHandler) {
			herr := h.Handle()
			if herr != nil {
				t.Error(herr)
			}
		}(handler)
	}
	for i := 0; i < nM; i++ {
		pNum := i % nP
		producers[pNum].Produce([]byte(fmt.Sprintf("message %d from producer %d", i, pNum)))
	}
	for i := 0; i < nM; i++ {
		fmt.Printf("Received: %s", <-msgChan)
	}
	for _, h := range handlers {
		err = h.Shutdown()
		if err != nil {
			t.Error(err)
		}
	}
	err = StopProducers()
	if err != nil {
		t.Error(err)
	}
}
