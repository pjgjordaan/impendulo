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

func SetAMQP_URI(uri string) {
	amqpURI = uri
}

func handleFunc(mh *MessageHandler) {
	herr := mh.Handle()
	if herr != nil {
		util.Log(herr)
	}
}

func Reply(ch amqp.Channel, d amqp.Delivery, body []byte) error {
	pub := amqp.Publishing{
		CorrelationId: d.CorrelationId,
		DeliveryMode:  amqp.Persistent,
		ContentType:   "text/plain",
		Body:          body,
		Priority:      0,
	}
	return ch.Publish(d.Exchange, d.ReplyTo, true, false, pub)
}

//NewHandler
func NewHandler(amqpURI, exchange, exchangeType, queue, ctag string, consumer Consumer, keys ...string) (mh *MessageHandler, err error) {
	defer func() {
		if err != nil {
			if mh.ch != nil {
				mh.ch.Close()
			}
			if mh.conn != nil {
				mh.conn.Close()
			}
		}
	}()
	mh = &MessageHandler{
		exchange: exchange,
		tag:      ctag,
		Consumer: consumer,
	}
	mh.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return
	}
	mh.ch, err = mh.conn.Channel()
	if err != nil {
		return
	}
	err = mh.ch.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return
	}
	isUnique := queue == ""
	q, err := mh.ch.QueueDeclare(
		queue,     // name of the queue
		!isUnique, // durable
		isUnique,  // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return
	}
	mh.ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	mh.queue = q.Name
	for _, key := range keys {
		err = mh.ch.QueueBind(
			q.Name,   // name of the queue
			key,      // bindingKey
			exchange, // sourceExchange
			false,    // noWait
			nil,      // arguments
		)
		if err != nil {
			return
		}
	}
	return
}

func (mh *MessageHandler) Handle() (err error) {
	deliveries, err := mh.ch.Consume(
		mh.queue, // name
		mh.tag,   // Tag,
		false,    // noAck
		false,    // exclusive
		false,    // noLocal
		false,    // noWait
		nil,      // arguments
	)
	if err != nil {
		return
	}
	for d := range deliveries {
		cerr := mh.Consume(d, *mh.ch)
		if err != nil {
			util.Log(cerr)
		}
	}
	return
}

func (mh *MessageHandler) Shutdown() (err error) {
	err = mh.ch.Cancel(mh.tag, true)
	if err != nil {
		return
	}
	err = mh.conn.Close()
	if err != nil {
		return
	}
	util.Log("AMQP shutdown OK")
	return
}

func (s *Submitter) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
	req := new(Request)
	if err = json.Unmarshal(d.Body, &req); err != nil {
		d.Ack(false)
		return
	}
	s.requestChan <- req
	d.Ack(false)
	return
}

func (s *Starter) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
	d.Ack(false)
	err = Reply(ch, d, []byte(s.key))
	return
}

func (c *Changer) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
	defer func() {
		d.Ack(false)
	}()
	status := new(Status)
	if err = json.Unmarshal(d.Body, &status); err != nil {
		return
	}
	c.statusChan <- *status
	return
}

func (w *Waiter) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
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
	err = Reply(ch, d, []byte{})
	return
}

func (sl *Loader) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
	sl.statusChan <- Status{}
	s := <-sl.statusChan
	body, err := json.Marshal(s)
	if err != nil {
		body = []byte(err.Error())
	}
	d.Ack(false)
	perr := Reply(ch, d, body)
	if err == nil && perr != nil {
		err = perr
	}
	return
}

func (r *Redoer) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
	defer func() {
		d.Ack(false)
	}()
	val := string(d.Body)
	subId, err := util.ReadId(val)
	if err != nil {
		err = fmt.Errorf("Not a valid Id %s", val)
		return
	}
	matcher := bson.M{db.SUBID: subId}
	selector := bson.M{db.DATA: 0}
	files, err := db.Files(matcher, selector, db.TIME)
	if err != nil {
		return
	}
	sreq := &Request{
		FileId: subId,
		SubId:  subId,
		Type:   SUBMISSION_START,
	}
	r.requestChan <- sreq
	defer func() {
		sreq.Type = SUBMISSION_STOP
		r.requestChan <- sreq
	}()
	for _, f := range files {
		freq := &Request{
			FileId: f.Id,
			SubId:  subId,
			Type:   FILE_ADD,
		}
		r.requestChan <- freq
	}
	return
}

func NewRedoer(requestChan chan *Request) (*MessageHandler, error) {
	redoer := &Redoer{
		requestChan: requestChan,
	}
	return NewHandler(amqpURI, SUB_REDO, DIRECT, "", "", redoer, "")
}

func NewSubmitter(requestChan chan *Request) (submitter, starter *MessageHandler, err error) {
	key := bson.NewObjectId().String()
	su := &Submitter{
		requestChan: requestChan,
	}
	st := &Starter{
		key: key,
	}
	submitter, err = NewHandler(amqpURI, "submission_exchange", DIRECT, "", "", su, key)
	if err != nil {
		return
	}
	starter, err = NewHandler(amqpURI, "submission_exchange", DIRECT, "", "", st, "submission_key")
	if err != nil {
		return
	}
	return
}

func NewWaiter(statusChan chan Status) (*MessageHandler, error) {
	waiter := &Waiter{
		statusChan: statusChan,
	}
	return NewHandler(amqpURI, "wait_exchange", DIRECT, "wait_queue", "", waiter, "wait_request_key")
}

func NewChanger(statusChan chan Status) (*MessageHandler, error) {
	changer := &Changer{
		statusChan: statusChan,
	}
	return NewHandler(amqpURI, "change_exchange", FANOUT, "", "", changer, "change_key")
}

func NewLoader(statusChan chan Status) (*MessageHandler, error) {
	loader := &Loader{
		statusChan: statusChan,
	}
	return NewHandler(amqpURI, "status_exchange", DIRECT, "status_queue", "", loader, "status_request_key")
}
