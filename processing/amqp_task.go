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
	fmt.Print()
	producers = make(map[string]*Producer)
	rps = make(map[string]*ReceiveProducer)
}

func NewReceiveProducer(name, amqpURI, exchange, exchangeType, publishKey, bindingKey, ctag string) (rp *ReceiveProducer, err error) {
	var ok bool
	if rp, ok = rps[name]; ok {
		return
	}
	rp = &ReceiveProducer{
		tag:        ctag,
		bindingKey: bindingKey,
	}
	rp.Producer, err = NewProducer(name, amqpURI, exchange, exchangeType, publishKey)
	if err != nil {
		return
	}
	q, err := rp.ch.QueueDeclare(
		"",    // name of the queue
		false, // durable
		true,  // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return
	}
	rp.queue = q.Name
	rp.ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	err = rp.ch.QueueBind(
		q.Name,        // name of the queue
		rp.bindingKey, // bindingKey
		exchange,      // sourceExchange
		false,         // noWait
		nil,           // arguments
	)
	if err == nil {
		rps[name] = rp
	}
	return
}

func (rp *ReceiveProducer) ReceiveProduce(data []byte) (reply []byte, err error) {
	u4, err := uuid.NewV4()
	if err != nil {
		return
	}
	cid := u4.String()
	msgs, err := rp.ch.Consume(rp.queue, rp.tag, false, false, false, false, nil)
	if err != nil {
		return
	}
	err = rp.ch.Publish(
		rp.exchange,   // publish to an exchange
		rp.publishKey, // routing to 0 or more queues
		true,          // mandatory
		false,         // immediate
		amqp.Publishing{
			CorrelationId: cid,
			ContentType:   "text/plain",
			Body:          data,
			DeliveryMode:  amqp.Persistent, // 1=non-persistent, 2=persistent
			Priority:      0,               // 0-9
		},
	)
	if err != nil {
		return
	}
	var d amqp.Delivery
	for d = range msgs {
		if d.CorrelationId == cid {
			d.Ack(false)
			reply = d.Body
			break
		}
	}
	return
}

func NewProducer(name, amqpURI, exchange, exchangeType, publishKey string) (p *Producer, err error) {
	var ok bool
	if p, ok = producers[name]; ok {
		return
	}
	p = &Producer{
		publishKey: publishKey,
		exchange:   exchange,
	}
	p.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return
	}
	p.ch, err = p.conn.Channel()
	if err != nil {
		return
	}
	err = p.ch.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	if err == nil {
		producers[name] = p
	}
	return
}

func (p *Producer) Produce(data []byte) error {
	return p.ch.Publish(
		p.exchange,   // publish to an exchange
		p.publishKey, // routing to 0 or more queues
		true,         // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         data,
			DeliveryMode: amqp.Persistent, // 1=non-persistent, 2=persistent
			Priority:     0,               // 0-9
		},
	)
}

func (p *Producer) Shutdown() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error: %s during producer shutdown on exchange %s.", err, p.exchange)
		}
	}()
	if p == nil {
		return
	}
	if p.ch != nil {
		err = p.ch.Close()
		if err != nil {
			return
		}
	}
	if p.conn != nil {
		err = p.conn.Close()
	}
	return
}

func StopProducers() (err error) {
	for _, p := range producers {
		if p == nil {
			continue
		}
		err = p.Shutdown()
		if err != nil {
			return
		}
	}
	producers = make(map[string]*Producer)
	rps = make(map[string]*ReceiveProducer)
	return
}

func StatusChanger(amqpURI string) (*Producer, error) {
	return NewProducer("status_changer", amqpURI, "change_exchange", FANOUT, "change_key")
}

//ChangeStatus is used to update Impendulo's current
//processing status.
func ChangeStatus(change Status) (err error) {
	sc, err := StatusChanger(AMQP_URI)
	if err != nil {
		return
	}
	marshalled, err := json.Marshal(change)
	if err != nil {
		return
	}
	err = sc.Produce(marshalled)
	return
}

//
func IdleWaiter(amqpURI string) (*ReceiveProducer, error) {
	return NewReceiveProducer("idle_waiter", amqpURI, "wait_exchange", DIRECT, "wait_request_key", "wait_response_key", "")
}

func WaitIdle() (err error) {
	idleWaiter, err := IdleWaiter(AMQP_URI)
	if err != nil {
		return
	}
	_, err = idleWaiter.ReceiveProduce(nil)
	return
}

func StatusRetriever(amqpURI string) (*ReceiveProducer, error) {
	return NewReceiveProducer("status_retriever", amqpURI, "status_exchange", DIRECT, "status_request_key", "status_response_key", "")
}

func GetStatus() (ret *Status, err error) {
	statusRetriever, err := StatusRetriever(AMQP_URI)
	if err != nil {
		return
	}
	resp, err := statusRetriever.ReceiveProduce(nil)
	if err != nil {
		return
	}
	ret = new(Status)
	err = json.Unmarshal(resp, &ret)
	return
}

func FileProducer(amqpURI string) (*Producer, error) {
	return NewProducer("file_producer", amqpURI, "file_exchange", DIRECT, "file_key")
}

func AddFile(file *project.File) (err error) {
	fileProducer, err := FileProducer(AMQP_URI)
	if err != nil {
		return
	}
	//We only need to process source files  and archives.
	if !file.CanProcess() {
		return nil
	}
	req := &Request{
		FileId: file.Id,
		SubId:  file.SubId,
	}
	marshalled, err := json.Marshal(req)
	if err != nil {
		return
	}
	err = fileProducer.Produce(marshalled)
	return
}

//EndProducer creates a new Producer which is used to signal the end of a submission.
func EndProducer(amqpURI string) (*Producer, error) {
	return NewProducer("end_producer", amqpURI, "end_exchange", FANOUT, "end_key")
}

//EndSubmission sends a message on AMQP that this submission has been completed by the user
//and can thus be closed when processing is done.
func EndSubmission(id bson.ObjectId) (err error) {
	endProducer, err := EndProducer(AMQP_URI)
	if err != nil {
		return
	}
	req := &Request{
		FileId: bson.NewObjectId(),
		SubId:  id,
		Stop:   true,
	}
	marshalled, err := json.Marshal(req)
	if err != nil {
		return
	}
	err = endProducer.Produce(marshalled)
	return
}
