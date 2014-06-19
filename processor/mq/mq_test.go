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

package mq

import (
	"fmt"

	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/status"
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

func th(mh *MessageHandler, t *testing.T) {
	if e := mh.Handle(); e != nil {
		t.Error(e)
	}
}

func init() {
	fmt.Sprint(time.Now(), project.Project{}, strconv.Itoa(1), bson.NewObjectId())
	util.SetErrorLogging("a")
	util.SetInfoLogging("f")
}

func (bc *BasicConsumer) Consume(d amqp.Delivery, ch *amqp.Channel) error {
	bc.msgs <- fmt.Sprintf("Consumer %d says %s.\n", bc.id, string(d.Body))
	d.Ack(false)
	return nil
}

func basicStatus() (status.S, bson.ObjectId) {
	sid := bson.NewObjectId().Hex()
	sm := map[string]map[string]util.E{
		sid: {
			bson.NewObjectId().Hex(): util.E{},
			bson.NewObjectId().Hex(): util.E{},
			bson.NewObjectId().Hex(): util.E{},
			bson.NewObjectId().Hex(): util.E{},
			bson.NewObjectId().Hex(): util.E{},
		},
	}
	return status.S{FileCount: len(sm[sid]), Submissions: sm}, bson.ObjectIdHex(sid)
}

func TestWaitIdle(t *testing.T) {
	wChan := make(chan util.E)
	w, e := NewWaiter(wChan)
	if e != nil {
		t.Error(e)
	}
	go th(w, t)
	s, sid := basicStatus()
	go func() {
		for s.FileCount > 0 {
			<-wChan
			wChan <- util.E{}
			for k, _ := range s.Submissions[sid.Hex()] {
				s.Update(request.RemoveFile(bson.ObjectIdHex(k), sid))
				break
			}
		}
		s.Update(request.StopSubmission(sid))
		<-wChan
		wChan <- util.E{}
	}()
	if e = WaitIdle(); e != nil {
		t.Error(e)
	}
	if e = w.Shutdown(); e != nil {
		t.Error(e)
	}
	if e = StopProducers(); e != nil {
		t.Error(e)
	}
}

func TestGetStatus(t *testing.T) {
	statusChan := make(chan status.S)
	sl, e := NewLoader(statusChan)
	if e != nil {
		t.Error(e)
	}
	go th(sl, t)
	status, _ := basicStatus()
	go func() {
		<-statusChan
		statusChan <- status
	}()
	s, e := GetStatus()
	if e != nil {
		t.Error(e)
	} else if s.FileCount != status.FileCount || len(s.Submissions) != len(status.Submissions) {
		t.Error(fmt.Errorf("invalid status %q", s))
	}
	if e = sl.Shutdown(); e != nil {
		t.Error(e)
	}
	if e = StopProducers(); e != nil {
		t.Error(e)
	}
}

func TestSubmitter(t *testing.T) {
	n := 100
	requestChan := make(chan *request.R)
	handlers := make([]*MessageHandler, 2*n)
	var e error
	for i := 0; i < n; i++ {
		if handlers[2*i+1], handlers[2*i], e = NewSubmitter(requestChan); e != nil {
			t.Error(e)
		}
	}
	for _, h := range handlers {
		go H(h)
	}
	time.Sleep(1 * time.Second)
	go func() {
		processed := 0
		for r := range requestChan {
			if r.Type != request.SUBMISSION_START {
				t.Error(fmt.Errorf("Invalid request %q.", r))
			}
			processed++
			if processed == n {
				break
			}
		}
	}()
	ids := make([]bson.ObjectId, n)
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = bson.NewObjectId()
		if keys[i], e = StartSubmission(ids[i]); e != nil {
			t.Error(e)
		}
	}
	time.Sleep(1 * time.Second)
	go func() {
		processed := 0
		for r := range requestChan {
			if r.Type != request.SUBMISSION_STOP {
				t.Error(fmt.Errorf("Invalid request %q.", r))
			}
			processed++
			if processed == n {
				break
			}
		}
	}()
	for i, id := range ids {
		if e = EndSubmission(id, keys[i]); e != nil {
			t.Error(e)
		}
	}
	for _, h := range handlers {
		if e = h.Shutdown(); e != nil {
			t.Error(e)
		}
	}
	if e = StopProducers(); e != nil {
		t.Error(e)
	}
}

func TestStatusChange(t *testing.T) {
	n := 10
	rChan := make(chan *request.R)
	handlers := make([]*MessageHandler, n)
	var e error
	for i := 0; i < n; i++ {
		if handlers[i], e = NewChanger(rChan); e != nil {
			t.Error(e)
		}
	}
	for _, h := range handlers {
		go func(fc *MessageHandler) {
			if e = fc.Handle(); e != nil {
				t.Error(e)
			}
		}(h)
	}
	r := request.StartSubmission(bson.NewObjectId())
	if e = ChangeStatus(r); e != nil {
		t.Error(e)
	}
	processed := 0
	for cur := range rChan {
		if *cur != *r {
			t.Error(fmt.Errorf("invalid change request %q", cur))
		}
		processed++
		if processed == n {
			break
		}
	}
	for _, h := range handlers {
		if e = h.Shutdown(); e != nil {
			t.Error(e)
		}
	}
	if e = StopProducers(); e != nil {
		t.Error(e)
	}
}

func TestAMQPBasic(t *testing.T) {
	msgChan := make(chan string)
	handler, e := NewHandler(DEFAULT_AMQP_URI, "test", DIRECT, "", "", &BasicConsumer{id: 1, msgs: msgChan}, "")
	if e != nil {
		t.Error(e)
	}
	producer, e := NewProducer("test_producer", DEFAULT_AMQP_URI, "test", DIRECT, "")
	if e != nil {
		t.Error(e)
	}
	go th(handler, t)
	n := 10
	for i := 0; i < n; i++ {
		producer.Produce([]byte(fmt.Sprintf("testing %d", i)))
	}
	for i := 0; i < n; i++ {
		fmt.Printf("Message %d %s", i, <-msgChan)
	}
	if e = handler.Shutdown(); e != nil {
		t.Error(e)
	}
	if e = StopProducers(); e != nil {
		t.Error(e)
	}
}

func TestAMQPQueue(t *testing.T) {
	nP, nH, nM := 10, 10, 50
	msgChan := make(chan string)
	handlers := make([]*MessageHandler, nH)
	var e error
	for i := 0; i < nH; i++ {
		if handlers[i], e = NewHandler(DEFAULT_AMQP_URI, "test", DIRECT, "test_queue", "", &BasicConsumer{id: i, msgs: msgChan}, "test_key"); e != nil {
			t.Error(e)
		}
	}
	producers := make([]*Producer, nP)
	for i := 0; i < nP; i++ {
		if producers[i], e = NewProducer("test_producer_"+strconv.Itoa(i), DEFAULT_AMQP_URI, "test", DIRECT, "test_key"); e != nil {
			t.Error(e)
		}
	}
	for _, handler := range handlers {
		go th(handler, t)
	}
	for i := 0; i < nM; i++ {
		pNum := i % nP
		producers[pNum].Produce([]byte(fmt.Sprintf("message %d from producer %d", i, pNum)))
	}
	for i := 0; i < nM; i++ {
		fmt.Printf("Received: %s", <-msgChan)
	}
	for _, h := range handlers {
		if e = h.Shutdown(); e != nil {
			t.Error(e)
		}
	}
	if e = StopProducers(); e != nil {
		t.Error(e)
	}
}
