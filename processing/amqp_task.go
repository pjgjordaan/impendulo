package processing

import (
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/nu7hatch/gouuid"
	"github.com/streadway/amqp"
	"labix.org/v2/mgo/bson"
)

const (
	TASK_QUEUE = "task_queue"
)

func Send(mId string, data []byte) (resp []byte, tipe string, err error) {
	conn, err := amqp.Dial(AMQP_URI)
	if err != nil {
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		TASK_QUEUE,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return
	}
	u4, err := uuid.NewV4()
	if err != nil {
		return
	}
	cId := u4.String()
	err = ch.Publish(
		"",           // exchange
		WORKER_QUEUE, // routing key
		true,         // mandatory
		false,
		amqp.Publishing{
			MessageId:     mId,
			CorrelationId: cId,
			ReplyTo:       q.Name,
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			Body:          data,
		})
	if err != nil {
		return
	}
	var d amqp.Delivery
	for d = range msgs {
		if d.CorrelationId == cId {
			break
		}
	}
	d.Ack(false)
	resp, tipe = d.Body, d.MessageId
	return
}

func RedoSubmission(id bson.ObjectId) (err error) {
	return sendId(id, SUB_REDO)
}

func StartSubmission(id bson.ObjectId) error {
	return sendId(id, SUB_START)
}

func EndSubmission(id bson.ObjectId) error {
	return sendId(id, SUB_END)
}

func sendId(id bson.ObjectId, tipe string) (err error) {
	resp, tipe, err := Send(tipe, []byte(id.Hex()))
	if err != nil {
		return
	}
	switch tipe {
	case SUCCESS:
	default:
		err = fmt.Errorf("Encountered error %s of type %s", string(resp), tipe)
	}
	return
}

func AddFile(file *project.File) (err error) {
	//We only need to process source files  and archives.
	if !file.CanProcess() {
		return nil
	}
	return sendId(file.Id, FILE)
}

func GetStatus() (ret *Status, err error) {
	ret = new(Status)
	resp, tipe, err := Send(STATUS, nil)
	if err != nil {
		return
	}
	switch tipe {
	case SUCCESS:
		err = json.Unmarshal(resp, &ret)
	default:
		err = fmt.Errorf("Encountered error %s of type %s", string(resp), tipe)
	}
	return
}

func WaitIdle() (err error) {
	_, _, err = Send(IDLE, nil)
	return
}
