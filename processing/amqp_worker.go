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

const (
	AMQP_URI                        = "amqp://guest:guest@localhost:5672/"
	LOG_AMQPWORKER                  = "processing/amqp_worker.go"
	WORKER_QUEUE                    = "worker_queue"
	SUB_START, SUB_END, SUB_REDO    = "submission_start", "submission_end", "submission_redo"
	FILE, STATUS                    = "file", "status"
	SUCCESS, IDLE                   = "success", "wait_idle"
	ERR_ID, ERR_REQUEST, ERR_STATUS = "error_id", "error_request", "error_status"
	PREFETCH_COUNT                  = 3
	PREFETCH_SIZE                   = 0
)

func startMQ(uri string, subCh, fileCh chan *Request, statusCh chan Status) (err error) {
	conn, err := amqp.Dial(uri)
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
		WORKER_QUEUE, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return
	}
	ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return
	}
	for d := range msgs {
		var body []byte
		var tipe string
		switch d.MessageId {
		case SUB_REDO:
			body, tipe = redoSubmission(subCh, fileCh, string(d.Body))
		case SUB_START, SUB_END:
			body, tipe = sendRequest(subCh, string(d.Body), d.MessageId == SUB_START)
		case FILE:
			body, tipe = sendRequest(fileCh, string(d.Body), false)
		case STATUS:
			body, tipe = getStatus(statusCh)
		case IDLE:
			waitIdle(statusCh)
		}
		d.Ack(false)
		pub := amqp.Publishing{
			MessageId:     tipe,
			CorrelationId: d.CorrelationId,
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			Body:          body,
		}
		perr := ch.Publish("", d.ReplyTo, true, false, pub)
		if perr != nil {
			util.Log(perr)
		}
	}
	return
}

func sendRequest(ch chan *Request, val string, start bool) (body []byte, tipe string) {
	if !bson.IsObjectIdHex(val) {
		body = []byte(fmt.Sprintf("Not a valid Id %s", val))
		tipe = ERR_ID
		return
	}
	req := &Request{
		Id:       bson.ObjectIdHex(val),
		Start:    start,
		Response: make(chan error),
	}
	ch <- req
	err := <-req.Response
	if err != nil {
		body = []byte(err.Error())
		tipe = ERR_REQUEST
		return
	}
	tipe = SUCCESS
	return
}

func waitIdle(statusCh chan Status) {
	idle := false
	for !idle {
		statusCh <- Status{}
		s := <-statusCh
		idle = s.Submissions == 0
		if !idle {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func getStatus(statusCh chan Status) (body []byte, tipe string) {
	statusCh <- Status{}
	s := <-statusCh
	body, jerr := json.Marshal(s)
	if jerr != nil {
		body = []byte(jerr.Error())
		tipe = ERR_STATUS
		return
	}
	tipe = SUCCESS
	return
}

func redoSubmission(subCh, fileCh chan *Request, val string) (body []byte, tipe string) {
	subId, err := util.ReadId(val)
	if err != nil {
		body = []byte(fmt.Sprintf("Not a valid Id %s", val))
		tipe = ERR_ID
		return
	}
	err = _redoSubmission(subCh, fileCh, subId)
	if err != nil {
		body = []byte(err.Error())
		tipe = ERR_REQUEST
		return
	}
	tipe = SUCCESS
	return
}

func _redoSubmission(subCh, fileCh chan *Request, subId bson.ObjectId) (err error) {
	req := &Request{
		Id:       subId,
		Response: make(chan error),
		Start:    true,
	}
	subCh <- req
	err = <-req.Response
	if err != nil {
		return
	}
	defer func() {
		req.Start = false
		subCh <- req
		serr := <-req.Response
		if err == nil && serr != nil {
			err = serr
		}
	}()
	matcher := bson.M{db.SUBID: subId}
	selector := bson.M{db.DATA: 0}
	files, err := db.Files(matcher, selector, db.TIME)
	if err != nil {
		return
	}
	for _, f := range files {
		freq := &Request{
			Id:       f.Id,
			Response: make(chan error),
		}
		fileCh <- freq
		err = <-freq.Response
		if err != nil {
			return
		}
	}
	return
}
