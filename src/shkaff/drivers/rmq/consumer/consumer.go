package consumer

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
	Msgs       <-chan amqp.Delivery
}

func InitAMQPConsumer(cfg config.ShkaffConfig) (qp *RMQ) {
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
	q, err := qp.Channel.QueueDeclare(
		"mongodb", // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalln(err, "Failed to declare a queue")
	}
	if err = qp.Channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		log.Fatalln(err, "Failed to set QoS")
	}
	if msgs, err := qp.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	); err != nil {
		log.Fatalln(err, "Failed to register a consumer")
	} else {
		qp.Msgs = msgs
	}

}
