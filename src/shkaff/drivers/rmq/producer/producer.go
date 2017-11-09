package producer

import (
	"fmt"
	"log"
	"shkaff/config"
	"shkaff/consts"

	"github.com/streadway/amqp"
)

type RMQ struct {
	uri        string
	queueName  string
	Channel    *amqp.Channel
	Connect    *amqp.Connection
	Publishing *amqp.Publishing
}

func InitAMQPProducer(cfg config.ShkaffConfig) (qp *RMQ) {
	qp = new(RMQ)
	qp.uri = fmt.Sprintf(consts.RMQ_URI_TEMPLATE, cfg.RMQ_USER,
		cfg.RMQ_PASS,
		cfg.RMQ_HOST,
		cfg.RMQ_PORT,
		cfg.RMQ_VHOST)
	qp.queueName = "mongodb"
	qp.initConnection()
	return
}

func (qp *RMQ) initConnection() {
	var err error
	if qp.Connect, err = amqp.Dial(qp.uri); err != nil {
		log.Fatalln(err)
	}
	if qp.Channel, err = qp.Connect.Channel(); err != nil {
		log.Fatalln(err)
	}
	if _, err = qp.Channel.QueueDeclare(
		qp.queueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		log.Fatalln(err)
	}
	qp.Publishing = new(amqp.Publishing)
	qp.Publishing.ContentType = "application/json"
}

func (qp *RMQ) Publish(body []byte) (err error) {
	qp.Publishing.Body = body
	if err = qp.Channel.Publish("", qp.queueName, false, false, *qp.Publishing); err != nil {
		return
	}
	return
}
