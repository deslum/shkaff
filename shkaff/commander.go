package shkaff

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	CONFIG_FILE       = "commander.json"
	DEFAULT_HOST      = "localhost"
	DEFAULT_AMQP_PORT = 5672

	URI_TEMPLATE = "%s://%s:%s@%s:%d/%s?sslmode=disable"
	AMQP         = "amqp"

	INVALID_AMQP_HOST     = "AMQP host in config file is empty. Shkaff set '%s'\n"
	INVALID_AMQP_PORT     = "AMPQ port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_AMQP_USER     = "AMQP user name is empty"
	INVALID_AMQP_PASSWORD = "AMQP password is empty"
)

var (
	opCache []Task
)

type ControlConfig struct {
	RMQ_HOST              string `json:"RMQ_HOST"`
	RMQ_PORT              int    `json:"RMQ_PORT"`
	RMQ_USER              string `json:"RMQ_USER"`
	RMQ_PASS              string `json:"RMQ_PASS"`
	RMQ_VHOST             string `json:"RMQ_VHOST"`
	DATABASE_HOST         string `json:"DATABASE_HOST"`
	DATABASE_PORT         int    `json:"DATABASE_PORT"`
	DATABASE_USER         string `json:"DATABASE_USER"`
	DATABASE_PASS         string `json:"DATABASE_PASS"`
	DATABASE_DB           string `json:"DATABASE_DB"`
	DATABASE_SSL          bool   `json:"DATABASE_SSL"`
	REFRESH_DATABASE_SCAN int    `json:"REFRESH_DATABASE_SCAN"`
}

type pSQL struct {
	uri             string
	refreshTimeScan int
}

type rmq struct {
	uri string
}

type Task struct {
	TaskID      int       `json:"task_id" db:"task_id"`
	Databases   string    `json:"-" db:"databases"`
	DBType      string    `json:"-" db:"db_type"`
	Verb        int       `json:"verb" db:"verb"`
	ThreadCount int       `json:"thread_count" db:"thread_count"`
	Gzip        bool      `json:"gzip" db:"gzip"`
	Ipv6        bool      `json:"ipv6" db:"ipv6"`
	Host        string    `json:"host" db:"host"`
	Port        int       `json:"port" db:"port"`
	StartTime   time.Time `json:"-" db:"start_time"`
	DBUser      string    `json:"db_user" db:"db_user"`
	DBPassword  string    `json:"db_password" db:"db_password"`
	Database    string    `json:"database"`
	Sheet       string    `json:"sheet"`
}

func initControlConfig(filename string) (cc ControlConfig) {
	var file []byte
	var err error
	if file, err = ioutil.ReadFile(filename); err != nil {
		log.Fatalln(err)
		return
	}
	if err := json.Unmarshal(file, &cc); err != nil {
		log.Fatalln(err)
		return
	}
	return
}

func (cc *ControlConfig) validateConfig() {

	if cc.RMQ_HOST == "" {
		log.Printf(INVALID_AMQP_HOST, DEFAULT_HOST)
		cc.RMQ_HOST = DEFAULT_HOST
	}
	if cc.RMQ_PORT < 1025 || cc.RMQ_PORT > 65535 {
		log.Printf(INVALID_AMQP_PORT, cc.RMQ_PORT, DEFAULT_AMQP_PORT)
		cc.RMQ_PORT = DEFAULT_AMQP_PORT
	}
	if cc.RMQ_USER == "" {
		log.Fatalln(INVALID_AMQP_USER)
	}
	if cc.RMQ_PASS == "" {
		log.Fatalln(INVALID_AMQP_PASSWORD)
	}
	return
}

func (qp *rmq) Connect() (channel *amqp.Channel) {
	var err error
	log.Println(qp.uri)
	conn, err := amqp.Dial(qp.uri)
	if err != nil {
		log.Fatalln("Connection", err)
	}
	channel, err = conn.Channel()
	if err != nil {
		log.Fatalln("Channel", err)
	}
	return channel

}

func initAMQP(cf ControlConfig) (qp *rmq) {
	qp = new(rmq)
	qp.uri = fmt.Sprintf(URI_TEMPLATE, AMQP,
		cf.RMQ_USER,
		cf.RMQ_PASS,
		cf.RMQ_HOST,
		cf.RMQ_PORT,
		cf.RMQ_VHOST)
	return
}

func main() {
	ch := make(chan bool)
	controlConfig := initControlConfig(CONFIG_FILE)
	controlConfig.validateConfig()
	qp := initAMQP(controlConfig)

	rmqChannel := qp.Connect()
	go func() {

		q, err := rmqChannel.QueueDeclare(
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

		if err = rmqChannel.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
		); err != nil {
			log.Fatalln(err, "Failed to set QoS")
		}

		msgs, err := rmqChannel.Consume(
			q.Name,      // queue
			"Commander", // consumer
			false,       // auto-ack
			false,       // exclusive
			false,       // no-local
			false,       // no-wait
			nil,         // args
		)

		if err != nil {
			log.Fatalln(err, "Failed to register a consumer")
		}
		var task Task
		for {
			for message := range msgs {
				if err := json.Unmarshal(message.Body, &task); err != nil {
					log.Println(err, "Failed JSON parse")
				}
				
				message.Ack(false)
			}
		}
	}()
	<-ch
}
