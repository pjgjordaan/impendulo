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
	"fmt"

	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/processor/status"
	"github.com/godfried/impendulo/util"
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
	}

	Filer struct {
		requestChan chan *request.R
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
	HandlerArgs       struct {
		Exchange  ExchangeArgs
		Queue     QueueArgs
		URI, CTAG string
		Keys      []string
	}
	ExchangeArgs struct {
		Name, Type                            string
		Durable, AutoDelete, Internal, NoWait bool
	}

	QueueArgs struct {
		Name                                   string
		Durable, AutoDelete, Exclusive, NoWait bool
	}
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
	amqpURI         = DEFAULT_AMQP_URI
	defaultExchange = ExchangeArgs{
		Type: DIRECT, Durable: true, AutoDelete: false, Internal: false, NoWait: false,
	}
	defaultQueue = QueueArgs{
		Durable: true, AutoDelete: false, Exclusive: false, NoWait: false,
	}
	defaultArgs = HandlerArgs{
		Queue: defaultQueue, Exchange: defaultExchange, URI: amqpURI, CTAG: "", Keys: []string{""},
	}
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
func NewHandler(args HandlerArgs, consumer Consumer) (*MessageHandler, error) {
	if args.CTAG == "" {
		u4, e := uuid.NewV4()
		if e != nil {
			return nil, e
		}
		args.CTAG = u4.String()
	}
	c, e := amqp.Dial(args.URI)
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
		args.Exchange.Name,
		args.Exchange.Type,
		args.Exchange.Durable,
		args.Exchange.AutoDelete,
		args.Exchange.Internal,
		args.Exchange.NoWait,
		nil, // arguments
	); e != nil {
		return nil, e
	}
	q, e := ch.QueueDeclare(
		args.Queue.Name,
		args.Queue.Durable,
		args.Queue.AutoDelete,
		args.Queue.Exclusive,
		args.Queue.NoWait,
		nil,
	)
	if e != nil {
		return nil, e
	}
	ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	for _, k := range args.Keys {
		if e = ch.QueueBind(
			q.Name,
			k,
			args.Exchange.Name,
			args.Queue.NoWait,
			nil,
		); e != nil {
			return nil, e
		}
	}
	return &MessageHandler{
		conn:     c,
		ch:       ch,
		queue:    q.Name,
		exchange: args.Exchange.Name,
		tag:      args.CTAG,
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

func (f *Filer) Consume(d amqp.Delivery, ch *amqp.Channel) (e error) {
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
	fmt.Println("filer received", r)
	if e = r.Valid(); e != nil {
		return
	}
	f.requestChan <- r
	fmt.Println("filer sent", r)
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

func args(exchange, queue, key string) HandlerArgs {
	a := defaultArgs
	a.Queue.Name = queue
	a.Exchange.Name = exchange
	a.Keys[0] = key
	return a
}

func NewSubmitter(rc chan *request.R) (*MessageHandler, error) {
	return NewHandler(args(submitterNames.exchange, submitterNames.queue, submitterNames.requestKey), &Submitter{requestChan: rc})
}

func NewFiler(rc chan *request.R, sid bson.ObjectId) (*MessageHandler, error) {
	a := args(filerNames.exchange, filerNames.queue+sid.Hex(), filerNames.requestKey+sid.Hex())
	a.Queue.AutoDelete = true
	return NewHandler(a, &Filer{requestChan: rc})
}

func NewWaiter(c chan util.E) (*MessageHandler, error) {
	return NewHandler(args(waiterNames.exchange, waiterNames.queue, waiterNames.requestKey), &Waiter{idleChan: c})
}

func NewChanger(c chan *request.R) (*MessageHandler, error) {
	return NewHandler(args(changerNames.exchange, changerNames.queue, changerNames.requestKey), &Changer{requestChan: c})
}

func NewLoader(c chan status.S) (*MessageHandler, error) {
	return NewHandler(args(retrieverNames.exchange, retrieverNames.queue, retrieverNames.requestKey), &Loader{statusChan: c})
}
