package producer

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	URI_TEMPLATE = "amqp://%s:%s@%s:%d/%s"
)

type RMQ struct {
	uri        string
	queueName  string
	channel    *amqp.Channel
	connect    *amqp.Connection
	publishing *amqp.Publishing
}

func InitAMQPProducer(user, password, host string, port int, vhost, queueName string) (qp *RMQ) {
	qp = new(RMQ)
	qp.uri = fmt.Sprintf(URI_TEMPLATE, user, password, host, port, vhost)
	qp.queueName = queueName
	qp.InitConnection()
	return
}

func (qp *RMQ) InitConnection() {
	var err error
	if qp.connect, err = amqp.Dial(qp.uri); err != nil {
		log.Panicln(err)
	}
	if qp.channel, err = qp.connect.Channel(); err != nil {
		log.Panicln(err)
	}
	if _, err = qp.channel.QueueDeclare(
		qp.queueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		log.Fatalln(err)
	}
	qp.publishing = new(amqp.Publishing)
	qp.publishing.ContentType = "application/json"
}

func (qp *RMQ) Publish(body string) (err error) {
	qp.publishing.Body = []byte(body)
	if err = qp.channel.Publish("", qp.queueName, false, false, *qp.publishing); err != nil {
		return
	}
	return
}
