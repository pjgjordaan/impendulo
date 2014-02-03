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
	Consumer interface {
		Consume(amqp.Delivery, amqp.Channel) error
	}
	Changer struct {
		statusChan chan Status
	}
	RequestConsumer struct {
		requestChan chan *Request
	}
	Loader struct {
		statusChan chan Status
		publishKey string
	}
	Waiter struct {
		statusChan chan Status
		publishKey string
	}
	Redoer struct {
		requestChan chan *Request
	}
	MessageHandler struct {
		conn                 *amqp.Connection
		ch                   *amqp.Channel
		tag, queue, exchange string
		done                 chan error
		Consumer
	}
	NewStatusHandler  func(amqpURI string, statusChan chan Status) (*MessageHandler, error)
	NewRequestHandler func(amqpURI string, statusChan chan *Request) (*MessageHandler, error)
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
	DIRECT                          = "direct"
	FANOUT                          = "fanout"
)

func NewHandler(amqpURI, exchange, exchangeType, queue, key, ctag string, consumer Consumer) (mh *MessageHandler, err error) {
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
		done:     make(chan error),
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
	q, err := mh.ch.QueueDeclare(
		queue, // name of the queue
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return
	}
	mh.ch.Qos(PREFETCH_COUNT, PREFETCH_SIZE, false)
	mh.queue = q.Name
	err = mh.ch.QueueBind(
		q.Name,   // name of the queue
		key,      // bindingKey
		exchange, // sourceExchange
		false,    // noWait
		nil,      // arguments
	)
	return
}

func (mh *MessageHandler) Handle() (err error) {
	defer func() {
		mh.done <- err
	}()
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
	defer util.Log("AMQP shutdown OK")
	err = <-mh.done
	mh = nil
	return
}

func (rc *RequestConsumer) Consume(d amqp.Delivery, ch amqp.Channel) (err error) {
	defer func() {
		d.Ack(false)
	}()
	req := new(Request)
	if err = json.Unmarshal(d.Body, &req); err != nil {
		return
	}
	rc.requestChan <- req
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
	pub := amqp.Publishing{
		CorrelationId: d.CorrelationId,
		DeliveryMode:  amqp.Persistent,
		ContentType:   "text/plain",
		Priority:      0,
	}
	perr := ch.Publish(d.Exchange, w.publishKey, true, false, pub)
	if err == nil && perr != nil {
		err = perr
	}
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
	pub := amqp.Publishing{
		CorrelationId: d.CorrelationId,
		DeliveryMode:  amqp.Persistent,
		ContentType:   "text/plain",
		Body:          body,
		Priority:      0,
	}
	perr := ch.Publish(d.Exchange, sl.publishKey, true, false, pub)
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
	defer func() {
		req := &Request{
			SubId: subId,
			Stop:  true,
		}
		r.requestChan <- req
	}()
	matcher := bson.M{db.SUBID: subId}
	selector := bson.M{db.DATA: 0}
	files, err := db.Files(matcher, selector, db.TIME)
	if err != nil {
		return
	}
	for _, f := range files {
		freq := &Request{
			FileId: f.Id,
			SubId:  subId,
		}
		r.requestChan <- freq
	}
	return
}

func NewRedoer(amqpURI string, requestChan chan *Request) (*MessageHandler, error) {
	redoer := &Redoer{
		requestChan: requestChan,
	}
	return NewHandler(amqpURI, SUB_REDO, DIRECT, "", "", "", redoer)
}

func NewFileConsumer(amqpURI string, requestChan chan *Request) (*MessageHandler, error) {
	r := &RequestConsumer{
		requestChan: requestChan,
	}
	return NewHandler(amqpURI, "file_exchange", DIRECT, "file_queue", "file_key", "", r)
}

func NewEnder(amqpURI string, requestChan chan *Request) (*MessageHandler, error) {
	r := &RequestConsumer{
		requestChan: requestChan,
	}
	return NewHandler(amqpURI, "end_exchange", FANOUT, "", "end_key", "", r)
}

func NewWaiter(amqpURI string, statusChan chan Status) (*MessageHandler, error) {
	waiter := &Waiter{
		statusChan: statusChan,
		publishKey: "wait_response_key",
	}
	return NewHandler(amqpURI, "wait_exchange", DIRECT, "wait_queue", "wait_request_key", "", waiter)
}

func NewChanger(amqpURI string, statusChan chan Status) (*MessageHandler, error) {
	changer := &Changer{
		statusChan: statusChan,
	}
	return NewHandler(amqpURI, "change_exchange", FANOUT, "", "change_key", "", changer)
}

func NewLoader(amqpURI string, statusChan chan Status) (*MessageHandler, error) {
	loader := &Loader{
		statusChan: statusChan,
		publishKey: "status_response_key",
	}
	return NewHandler(amqpURI, "status_exchange", DIRECT, "status_queue", "status_request_key", "", loader)
}
