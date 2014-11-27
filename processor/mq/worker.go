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
	"encoding/json"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/status"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/streadway/amqp"
	"labix.org/v2/mgo/bson"
)

type (
	//Consumer is an interface for allowing the processing of messages from AMQP.
	Consumer interface {
		Consume(amqp.Delivery, *amqp.Channel) error
	}

	//Changer is a Consumer which listens for updates to Impendulo's status
	//and changes it accordingly.
	Changer struct {
		requestChan chan *request.R
	}

	//Submitter is a Consumer used to handle submission and file requests.
	Submitter struct {
		requestChan chan *request.R
		key         string
	}

	//Starter is a Consumer used to connect a Submitter to a submission request.
	Starter struct {
		key string
	}

	//Loader is a Consumer which listens for status requests on AMQP
	//and responds to them with Impendulo's current status.
	Loader struct {
		statusChan chan status.S
	}

	//Waiter is a Consumer which listens for requests for when Impendulo is idle
	//and responds to them when it is.
	Waiter struct {
		idleChan chan util.E
	}

	//Redoer is a Consumer which listens for requests to reanalyse submissions
	//and submits them for reanalysis.
	Redoer struct {
		requestChan chan *request.R
	}

	//MessageHandler wraps a consumer in a struct in order to provide with other
	//tools to manage its AMQP connection.
	MessageHandler struct {
		conn                 *amqp.Connection
		ch                   *amqp.Channel
		tag, queue, exchange string
		Consumer
	}

	NewStatusHandler func(amqpURI string, statusChan chan status.S) (*MessageHandler, error)

	NewRequestHandler func(amqpURI string, requestChan chan *request.R) (*MessageHandler, error)
)

const (
	DEFAULT_AMQP_URI                = "amqp://guest:guest@localhost:5672/"
	LOG_AMQPWORKER                  = "processing/amqp_worker.go"
	WORKER_QUEUE                    = "worker_queue"
	SUB_START, SUB_END, SUB_REDO    = "submission_start", "submission_end", "submission_redo"
	FILE, STATUS                    = "file", "status"
	SUCCESS, IDLE                   = "success", "wait_idle"
	ERR_ID, ERR_REQUEST, ERR_STATUS = "error_id", "error_request", "error_status"
	PREFETCH_COUNT                  = 3
	PREFETCH_SIZE                   = 0
	DIRECT                          = "direct"
	FANOUT                          = "fanout"
)

var (
	amqpURI = DEFAULT_AMQP_URI
)

func SetAMQP_URI(uri string) {
	amqpURI = uri
}

func H(m *MessageHandler) {
	if e := m.Handle(); e != nil {
		util.Log(e, m.Shutdown())
	}
}

func Reply(c *amqp.Channel, d amqp.Delivery, b []byte) error {
	return c.Publish(d.Exchange, d.ReplyTo, true, false,
		amqp.Publishing{
			CorrelationId: d.CorrelationId,
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			Body:          b,
			Priority:      0,
		})
}

//NewHandler
func NewHandler(amqpURI, exchange, exchangeType, queue, ctag string, consumer Consumer, keys ...string) (*MessageHandler, error) {
	if ctag == "" {
		u4, e := uuid.NewV4()
		if e != nil {
			return nil, e
		}
		ctag = u4.String()
	}
	c, e := amqp.Dial(amqpURI)
	if e != nil {
		return nil, e
	}
	conErrs := c.NotifyClose(make(chan *amqp.Error))
	go func() {
		for e := range conErrs {
			util.Log(e)
		}
	}()
	ch, e := c.Channel()
	if e != nil {
		return nil, e
	}
	chErrs := ch.NotifyClose(make(chan *amqp.Error))
	go func() {
		for e := range chErrs {
			util.Log(e)
		}
	}()
	if e = ch.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); e != nil {
		return nil, e
	}
	isUnique := queue == ""
	q, e := ch.QueueDeclare(
		queue,     // name of the queue
		!isUnique, // durable
		isUnique,  // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if e != nil {
		return nil, e
	}
	ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	for _, k := range keys {
		if e = ch.QueueBind(
			q.Name,   // name of the queue
			k,        // bindingKey
			exchange, // sourceExchange
			false,    // noWait
			nil,      // arguments
		); e != nil {
			return nil, e
		}
	}
	return &MessageHandler{
		conn:     c,
		ch:       ch,
		queue:    q.Name,
		exchange: exchange,
		tag:      ctag,
		Consumer: consumer,
	}, nil
}

func (m *MessageHandler) Handle() error {
	ds, e := m.ch.Consume(
		m.queue, // name
		m.tag,   // Tag,
		false,   // noAck
		false,   // exclusive
		false,   // noLocal
		false,   // noWait
		nil,     // arguments
	)
	if e != nil {
		return e
	}
	for d := range ds {
		if e := m.Consume(d, m.ch); e != nil {
			util.Log(e)
		}
	}
	return nil
}

func (m *MessageHandler) DeleteQueue() error {
	_, e := m.ch.QueueDelete(m.queue, false, false, false)
	return e
}

func (m *MessageHandler) Shutdown() error {
	if e := m.ch.Close(); e != nil {
		return e
	}
	if e := m.conn.Close(); e != nil {
		return e
	}
	util.Log("AMQP shutdown OK")
	return nil
}

func (s *Submitter) Consume(d amqp.Delivery, ch *amqp.Channel) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
		d.Ack(false)
	}()
	r := new(request.R)
	if e = json.Unmarshal(d.Body, &r); e != nil {
		return
	}
	if e = r.Valid(); e != nil {
		return
	}
	s.requestChan <- r
	return
}

func (c *Changer) Consume(d amqp.Delivery, ch *amqp.Channel) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
		d.Ack(false)
	}()
	r := new(request.R)
	if e = json.Unmarshal(d.Body, &r); e != nil {
		return
	}
	if e = r.Valid(); e != nil {
		return
	}
	c.requestChan <- r
	return
}

func (w *Waiter) Consume(d amqp.Delivery, ch *amqp.Channel) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
		d.Ack(false)
	}()
	w.idleChan <- util.E{}
	<-w.idleChan
	e = Reply(ch, d, []byte{})
	return
}

func (l *Loader) Consume(d amqp.Delivery, ch *amqp.Channel) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(error)
		}
		d.Ack(false)
	}()
	l.statusChan <- status.S{}
	s := <-l.statusChan
	b, e := json.Marshal(s)
	if e != nil {
		b = []byte(e.Error())
	}
	re := Reply(ch, d, b)
	if e == nil && re != nil {
		e = re
	}
	return
}

func (r *Redoer) Consume(d amqp.Delivery, ch *amqp.Channel) (e error) {
	defer func() {
		if rec := recover(); rec != nil {
			e = rec.(error)
		}
		d.Ack(false)
	}()
	sid, e := convert.Id(string(d.Body))
	if e != nil {
		return
	}
	fs, e := db.Files(bson.M{db.SUBID: sid}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return
	}
	r.requestChan <- request.StartSubmission(sid)
	for _, f := range fs {
		if !f.CanProcess() {
			continue
		}
		var req *request.R
		if req, e = request.AddFile(f); e != nil {
			return
		}
		r.requestChan <- req
	}
	r.requestChan <- request.StopSubmission(sid)
	return
}

func NewRedoer(rc chan *request.R) (*MessageHandler, error) {
	return NewHandler(amqpURI, "submission_exchange", DIRECT, "redo_queue", "", &Redoer{requestChan: rc}, "redo_key")
}

func NewSubmitter(rc chan *request.R) (*MessageHandler, error) {
	return NewHandler(amqpURI, "submission_exchange", DIRECT, "submission_queue", "", &Submitter{requestChan: rc}, "submission_key")
}

func NewFiler(rc chan *request.R, sid bson.ObjectId) (*MessageHandler, error) {
	return NewHandler(amqpURI, "submission_exchange", DIRECT, "file_queue_"+sid.Hex(), "", &Submitter{requestChan: rc}, "file_key_"+sid.Hex())
}

func NewWaiter(c chan util.E) (*MessageHandler, error) {
	return NewHandler(amqpURI, "wait_exchange", DIRECT, "wait_queue", "", &Waiter{idleChan: c}, "wait_request_key")
}

func NewChanger(c chan *request.R) (*MessageHandler, error) {
	return NewHandler(amqpURI, "change_exchange", FANOUT, "", "", &Changer{requestChan: c}, "change_key")
}

func NewLoader(c chan status.S) (*MessageHandler, error) {
	return NewHandler(amqpURI, "status_exchange", DIRECT, "status_queue", "", &Loader{statusChan: c}, "status_request_key")
}
