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
	"encoding/json"
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/streadway/amqp"
	"labix.org/v2/mgo/bson"

	"time"
)

type (
	//Consumer is an interface for allowing the processing of messages from AMQP.
	Consumer interface {
		Consume(amqp.Delivery, amqp.Channel) error
	}

	//Changer is a Consumer which listens for updates to Impendulo's status
	//and changes it accordingly.
	Changer struct {
		statusChan chan Status
	}

	//Submitter is a Consumer used to handle submission and file requests.
	Submitter struct {
		requestChan chan *Request
		key         string
	}

	//Starter is a Consumer used to connect a Submitter to a submission request.
	Starter struct {
		key string
	}

	//Loader is a Consumer which listens for status requests on AMQP
	//and responds to them with Impendulo's current status.
	Loader struct {
		statusChan chan Status
	}

	//Waiter is a Consumer which listens for requests for when Impendulo is idle
	//and responds to them when it is.
	Waiter struct {
		statusChan chan Status
	}

	//Redoer is a Consumer which listens for requests to reanalyse submissions
	//and submits them for reanalysis.
	Redoer struct {
		requestChan chan *Request
	}

	//MessageHandler wraps a consumer in a struct in order to provide with other
	//tools to manage its AMQP connection.
	MessageHandler struct {
		conn                 *amqp.Connection
		ch                   *amqp.Channel
		tag, queue, exchange string
		Consumer
	}

	NewStatusHandler func(amqpURI string, statusChan chan Status) (*MessageHandler, error)

	NewRequestHandler func(amqpURI string, statusChan chan *Request) (*MessageHandler, error)
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

func init() {
	fmt.Sprint()
}

func SetAMQP_URI(uri string) {
	amqpURI = uri
}

func handleFunc(m *MessageHandler) {
	if e := m.Handle(); e != nil {
		util.Log(e)
	}
}

func Reply(c amqp.Channel, d amqp.Delivery, b []byte) error {
	p := amqp.Publishing{
		CorrelationId: d.CorrelationId,
		DeliveryMode:  amqp.Persistent,
		ContentType:   "text/plain",
		Body:          b,
		Priority:      0,
	}
	return c.Publish(d.Exchange, d.ReplyTo, true, false, p)
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
	ch, e := c.Channel()
	if e != nil {
		return nil, e
	}
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
		e := m.Consume(d, *m.ch)
		if e != nil {
			util.Log(e)
		}
	}
	return nil
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

func (s *Submitter) Consume(d amqp.Delivery, ch amqp.Channel) error {
	defer func() {
		d.Ack(false)
	}()
	r := new(Request)
	if e := json.Unmarshal(d.Body, &r); e != nil {
		return e
	}
	s.requestChan <- r
	return nil
}

func (s *Starter) Consume(d amqp.Delivery, ch amqp.Channel) error {
	d.Ack(false)
	return Reply(ch, d, []byte(s.key))
}

func (c *Changer) Consume(d amqp.Delivery, ch amqp.Channel) error {
	defer func() {
		d.Ack(false)
	}()
	s := new(Status)
	if e := json.Unmarshal(d.Body, &s); e != nil {
		return e
	}
	c.statusChan <- *s
	return nil
}

func (w *Waiter) Consume(d amqp.Delivery, ch amqp.Channel) error {
	idle := false
	for !idle {
		w.statusChan <- Status{}
		s := <-w.statusChan
		idle = s.Submissions == 0
		if !idle {
			time.Sleep(100 * time.Millisecond)
		}
	}
	d.Ack(false)
	return Reply(ch, d, []byte{})
}

func (l *Loader) Consume(d amqp.Delivery, ch amqp.Channel) error {
	l.statusChan <- Status{}
	s := <-l.statusChan
	b, e := json.Marshal(s)
	if e != nil {
		b = []byte(e.Error())
	}
	d.Ack(false)
	re := Reply(ch, d, b)
	if e == nil && re != nil {
		e = re
	}
	return e
}

func (r *Redoer) Consume(d amqp.Delivery, ch amqp.Channel) error {
	defer func() {
		d.Ack(false)
	}()
	sid, e := util.ReadId(string(d.Body))
	if e != nil {
		return e
	}
	fs, e := db.Files(bson.M{db.SUBID: sid}, bson.M{db.DATA: 0}, db.TIME)
	if e != nil {
		return e
	}
	r.requestChan <- &Request{
		FileId: sid,
		SubId:  sid,
		Type:   SUBMISSION_START,
	}
	for _, f := range fs {
		if !f.CanProcess() {
			continue
		}
		r.requestChan <- &Request{
			FileId: f.Id,
			SubId:  sid,
			Type:   FILE_ADD,
		}
	}
	r.requestChan <- &Request{
		FileId: sid,
		SubId:  sid,
		Type:   SUBMISSION_STOP,
	}
	return nil
}

func NewRedoer(rc chan *Request) (*MessageHandler, error) {
	return NewHandler(amqpURI, "submission_exchange", DIRECT, "redo_queue", "", &Redoer{requestChan: rc}, "redo_key")
}

func NewSubmitter(rc chan *Request) (*MessageHandler, *MessageHandler, error) {
	k := bson.NewObjectId().String()
	su, e := NewHandler(amqpURI, "submission_exchange", DIRECT, "", "", &Submitter{requestChan: rc}, k)
	if e != nil {
		return nil, nil, e
	}
	st, e := NewHandler(amqpURI, "submission_exchange", DIRECT, "", "", &Starter{key: k}, "submission_key")
	if e != nil {
		return nil, nil, e
	}
	return su, st, nil
}

func NewWaiter(sc chan Status) (*MessageHandler, error) {
	return NewHandler(amqpURI, "wait_exchange", DIRECT, "wait_queue", "", &Waiter{statusChan: sc}, "wait_request_key")
}

func NewChanger(sc chan Status) (*MessageHandler, error) {
	return NewHandler(amqpURI, "change_exchange", FANOUT, "", "", &Changer{statusChan: sc}, "change_key")
}

func NewLoader(sc chan Status) (*MessageHandler, error) {
	return NewHandler(amqpURI, "status_exchange", DIRECT, "status_queue", "", &Loader{statusChan: sc}, "status_request_key")
}
