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
	"github.com/godfried/impendulo/project"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/streadway/amqp"
	"labix.org/v2/mgo/bson"
)

type (
	//Producer is used to create new tasks which it publishes to the queue.
	Producer struct {
		conn                 *amqp.Connection
		ch                   *amqp.Channel
		publishKey, exchange string
	}
	//ReceiveProducer is used to create new tasks which it publishes to the queue.
	//It also receives a response from the consumer which received its task.
	ReceiveProducer struct {
		tag, queue, bindingKey string
		*Producer
	}
)

var (
	producers map[string]*Producer
	rps       map[string]*ReceiveProducer
)

const (
	TASK_QUEUE = "task_queue"
)

func init() {
	producers = make(map[string]*Producer)
	rps = make(map[string]*ReceiveProducer)
}

//NewReceiveProducer
func NewReceiveProducer(name, amqpURI, exchange, exchangeType, publishKey, bindingKey, ctag string) (*ReceiveProducer, error) {
	if r, ok := rps[name]; ok {
		return r, nil
	}
	if ctag == "" {
		u4, e := uuid.NewV4()
		if e != nil {
			return nil, e
		}
		ctag = u4.String()
	}
	p, e := NewProducer(name, amqpURI, exchange, exchangeType, publishKey)
	if e != nil {
		return nil, e
	}
	q, e := p.ch.QueueDeclare(
		"",    // name of the queue
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if e != nil {
		return nil, e
	}
	p.ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	if e = p.ch.QueueBind(
		q.Name,     // name of the queue
		bindingKey, // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); e != nil {
		return nil, e
	}
	r := &ReceiveProducer{
		queue:      q.Name,
		tag:        ctag,
		bindingKey: bindingKey,
		Producer:   p,
	}
	rps[name] = r
	return r, nil
}

//ReceiveProduce
func (r *ReceiveProducer) ReceiveProduce(d []byte) ([]byte, error) {
	u4, e := uuid.NewV4()
	if e != nil {
		return nil, e
	}
	cid := u4.String()
	ds, e := r.ch.Consume(r.queue, r.tag, false, false, false, false, nil)
	if e != nil {
		return nil, e
	}
	if e = r.ch.Publish(
		r.exchange,   // publish to an exchange
		r.publishKey, // routing to 0 or more queues
		true,         // mandatory
		false,        // immediate
		amqp.Publishing{
			ReplyTo:       r.bindingKey,
			CorrelationId: cid,
			ContentType:   "text/plain",
			Body:          d,
			DeliveryMode:  amqp.Persistent, // 1=non-persistent, 2=persistent
			Priority:      0,               // 0-9
		},
	); e != nil {
		return nil, e
	}
	var reply []byte
	for d := range ds {
		if d.CorrelationId == cid {
			d.Ack(false)
			reply = d.Body
			break
		}
	}
	if e = r.ch.Cancel(r.tag, false); e != nil {
		return nil, e
	}
	return reply, nil
}

//NewProducer
func NewProducer(name, amqpURI, exchange, exchangeType, publishKey string) (*Producer, error) {
	if p, ok := producers[name]; ok {
		return p, nil
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
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); e != nil {
		return nil, e
	}
	p := &Producer{
		conn:       c,
		ch:         ch,
		publishKey: publishKey,
		exchange:   exchange,
	}
	producers[name] = p
	return p, nil
}

//Produce publishes the provided data on the amqp.Channel as configured previously.
func (p *Producer) Produce(d []byte) error {
	return p.ch.Publish(
		p.exchange,   // publish to an exchange
		p.publishKey, // routing to 0 or more queues
		true,         // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         d,
			DeliveryMode: amqp.Persistent, // 1=non-persistent, 2=persistent
			Priority:     0,               // 0-9
		},
	)
}

//Shutdown stops this Producer by closing its channel and connection.
func (p *Producer) Shutdown() error {
	if p.ch != nil {
		if e := p.ch.Close(); e != nil {
			return e
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

//StopProducers shuts all active producers down.
func StopProducers() error {
	for _, p := range producers {
		if p == nil {
			continue
		}
		if e := p.Shutdown(); e != nil {
			return e
		}
	}
	producers = make(map[string]*Producer)
	rps = make(map[string]*ReceiveProducer)
	return nil
}

//StatusChanger creates a Producer which can update Impendulo's status.
func StatusChanger(amqpURI string) (*Producer, error) {
	return NewProducer("status_changer", amqpURI, "change_exchange", FANOUT, "change_key")
}

//ChangeStatus is used to update Impendulo's current
//processing status.
func ChangeStatus(r Request) error {
	if e := r.Valid(); e != nil {
		return e
	}
	s, e := StatusChanger(amqpURI)
	if e != nil {
		return e
	}
	m, e := json.Marshal(r)
	if e != nil {
		return e
	}
	return s.Produce(m)
}

//IdleWaiter
func IdleWaiter(amqpURI string) (*ReceiveProducer, error) {
	return NewReceiveProducer("idle_waiter", amqpURI, "wait_exchange", DIRECT, "wait_request_key", "wait_response_key", "")
}

//WaitIdle will only return once impendulo's processors are idle when called.
func WaitIdle() error {
	w, e := IdleWaiter(amqpURI)
	if e != nil {
		return e
	}
	_, e = w.ReceiveProduce(nil)
	return e
}

//StatusRetriever
func StatusRetriever(amqpURI string) (*ReceiveProducer, error) {
	return NewReceiveProducer("status_retriever", amqpURI, "status_exchange", DIRECT, "status_request_key", "status_response_key", "")
}

//GetStatus retrieves the current status of impendulo's processors
func GetStatus() (*Status, error) {
	sr, e := StatusRetriever(amqpURI)
	if e != nil {
		return nil, e
	}
	r, e := sr.ReceiveProduce(nil)
	if e != nil {
		return nil, e
	}
	s := Status{}
	if e = json.Unmarshal(r, &s); e != nil {
		return nil, e
	}
	return &s, nil
}

//FileProducer
func FileProducer(amqpURI string, fileKey string) (*Producer, error) {
	return NewProducer("file_producer_"+fileKey, amqpURI, "submission_exchange", DIRECT, fileKey)
}

//AddFile
func AddFile(f *project.File, k string) error {
	p, e := FileProducer(amqpURI, k)
	if e != nil {
		return e
	}
	//We only need to process source files  and archives.
	if !f.CanProcess() {
		return nil
	}
	m, e := json.Marshal(Request{
		FileId: f.Id,
		SubId:  f.SubId,
		Type:   FILE_ADD,
	})
	if e != nil {
		return e
	}
	return p.Produce(m)
}

//StartProducer creates a new Producer which is used to signal the start or end of a submission.
func StartProducer(amqpURI string) (*ReceiveProducer, error) {
	id := bson.NewObjectId().String()
	return NewReceiveProducer("submission_producer_"+id, amqpURI, "submission_exchange", DIRECT, "submission_key", id, "")
}

//StartSubmission
func StartSubmission(id bson.ObjectId) (string, error) {
	p, e := StartProducer(amqpURI)
	if e != nil {
		return "", e
	}
	m, e := json.Marshal(Request{
		FileId: id,
		SubId:  id,
		Type:   SUBMISSION_START,
	})
	if e != nil {
		return "", e
	}
	d, e := p.ReceiveProduce([]byte{})
	if e != nil {
		return "", e
	}
	p.publishKey = string(d)
	if e = p.Produce(m); e != nil {
		return "", e
	}
	return p.publishKey, nil
}

//EndSubmission sends a message on AMQP that this submission has been completed by the user
//and can thus be closed when processing is done.
func EndSubmission(id bson.ObjectId, k string) error {
	p, e := FileProducer(amqpURI, k)
	if e != nil {
		return e
	}
	m, e := json.Marshal(Request{
		FileId: id,
		SubId:  id,
		Type:   SUBMISSION_STOP,
	})
	if e != nil {
		return e
	}
	return p.Produce(m)
}

//RedoProducer
func RedoProducer(amqpURI string) (*Producer, error) {
	return NewProducer("redo_producer", amqpURI, "submission_exchange", DIRECT, "redo_key")
}

//RedoSubmission
func RedoSubmission(id bson.ObjectId) error {
	p, e := RedoProducer(amqpURI)
	if e != nil {
		return e
	}
	return p.Produce([]byte(id.Hex()))
}
