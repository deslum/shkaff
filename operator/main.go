package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/streadway/amqp"

	_ "github.com/lib/pq"
)

const (
	CONFIG_FILE           = "operator.json"
	DEFAULT_HOST          = "localhost"
	DEFAULT_DATABASE_PORT = 5432
	DEFAULT_AMQP_PORT     = 5672
	DEFAULT_DATABASE_DB   = "postgres"

	URI_TEMPLATE = "%s://%s:%s@%s:%d/%s?sslmode=disable"
	POSTGRES     = "postgres"
	AMQP         = "amqp"

	INVALID_DATABASE_HOST     = "Database host in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_PORT     = "Database port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_DATABASE_DB       = "Database name in config file is empty. Shkaff set '%s'\n"
	INVALID_DATABASE_USER     = "Database user name is empty"
	INVALID_DATABASE_PASSWORD = "Database password is empty"

	INVALID_AMQP_HOST     = "AMQP host in config file is empty. Shkaff set '%s'\n"
	INVALID_AMQP_PORT     = "AMPQ port %d in config file invalid. Shkaff set '%d'\n"
	INVALID_AMQP_USER     = "AMQP user name is empty"
	INVALID_AMQP_PASSWORD = "AMQP password is empty"

	REQUEST_GET_STARTTIME = "SELECT task_name FROM shkaff.tasks WHERE start_time > to_timestamp(%d)"

	REFRESH_DATABASE_SCAN = 5
	TASKS_TIME            = 10
)

type ControlConfig struct {
	RMQ_HOST      string `json:"RMQ_HOST"`
	RMQ_PORT      int    `json:"RMQ_PORT"`
	RMQ_USER      string `json:"RMQ_USER"`
	RMQ_PASS      string `json:"RMQ_PASS"`
	RMQ_VHOST     string `json:"RMQ_VHOST"`
	DATABASE_HOST string `json:"DATABASE_HOST"`
	DATABASE_PORT int    `json:"DATABASE_PORT"`
	DATABASE_USER string `json:"DATABASE_USER"`
	DATABASE_PASS string `json:"DATABASE_PASS"`
	DATABASE_DB   string `json:"DATABASE_DB"`
	DATABASE_SSL  bool   `json:"DATABASE_SSL"`
}

type pSQL struct {
	uri string
}

type rmq struct {
	uri string
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
	if cc.DATABASE_HOST == "" {
		log.Printf(INVALID_DATABASE_HOST, DEFAULT_HOST)
		cc.DATABASE_HOST = DEFAULT_HOST
	}
	if cc.DATABASE_PORT < 1025 || cc.DATABASE_PORT > 65535 {
		log.Printf(INVALID_DATABASE_PORT, cc.DATABASE_PORT, DEFAULT_DATABASE_PORT)
		cc.DATABASE_PORT = DEFAULT_DATABASE_PORT
	}
	if cc.DATABASE_DB == "" {
		log.Printf(INVALID_DATABASE_DB, DEFAULT_DATABASE_DB)
		cc.DATABASE_DB = DEFAULT_DATABASE_DB
	}
	if cc.DATABASE_USER == "" {
		log.Fatalln(INVALID_DATABASE_USER)
	}
	if cc.DATABASE_PASS == "" {
		log.Fatalln(INVALID_DATABASE_PASSWORD)
	}

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

func initPSQL(cf ControlConfig) (ps *pSQL) {
	ps = new(pSQL)
	ps.uri = fmt.Sprintf(URI_TEMPLATE, POSTGRES,
		cf.DATABASE_USER,
		cf.DATABASE_PASS,
		cf.DATABASE_HOST,
		cf.DATABASE_PORT,
		cf.DATABASE_DB)
	return
}
func (ps *pSQL) Connect() (db *sql.DB) {
	var err error
	log.Println(ps.uri)
	db, err = sql.Open(POSTGRES, ps.uri)
	if err != nil {
		log.Fatalln(err)
	}
	return
}

func (qp *rmq) Connect() (channel *amqp.Channel) {
	var err error
	log.Println(qp.uri)
	conn, err := amqp.Dial(qp.uri)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	channel, err = conn.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	defer channel.Close()
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

type Task struct {
	taskid       int
	task_name    string
	verbose      int
	start_time   int64
	is_active    bool
	thread_count int
}

var opCache []Task

func isDublicateTask(opc []Task, task Task) (result bool) {
	for _, oc := range opc {
		if oc.taskid == task.taskid {
			return true
		}
	}
	return false
}

func remove(slice []Task, s int) []Task {
	return append(slice[:s], slice[s+1:]...)
}

func main() {
	var task Task
	ch := make(chan bool)
	var psqlUpdateTime *time.Timer
	controlConfig := initControlConfig(CONFIG_FILE)
	controlConfig.validateConfig()
	pSQL := initPSQL(controlConfig)
	qp := initAMQP(controlConfig)
	db := pSQL.Connect()
	rmqChannel := qp.Connect()
	//Aggregator
	psqlUpdateTime = time.NewTimer(REFRESH_DATABASE_SCAN * time.Second)

	go func() {
		for {
			select {
			case <-psqlUpdateTime.C:
				request := fmt.Sprintf(REQUEST_GET_STARTTIME, time.Now().Unix()+TASKS_TIME)
				rows, err := db.Query(request)
				defer rows.Close()
				if err != nil {
					log.Println(err)
				} else {
					for rows.Next() {
						if err := rows.Scan(&task); err != nil {
							log.Println(err)
						}
						if !isDublicateTask(opCache, task) {
							opCache = append(opCache, task)
						}
					}
				}
				psqlUpdateTime = time.NewTimer(REFRESH_DATABASE_SCAN * time.Second)
			}
		}
	}()

	//TaskSender
	go func() {
		for {
			for numEl, cache := range opCache {
				queue, err := rmqChannel.QueueDeclare(
					"for_worker", //name
					true,         // durable
					false,        // delete when unused
					false,        // exclusive
					false,        // no-wait
					nil,          // arguments
				)
				if err != nil {
					log.Println(err)
				}
				if cache.start_time > time.Now().Unix() {
					body, err := json.Marshal(cache)
					if err != nil {
						log.Println(err)
						continue
					}
					pub := amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(body),
					}
					if err := rmqChannel.Publish("", queue.Name, false, false, pub); err != nil {
						log.Println(err)
					} else {
						opCache = remove(opCache, numEl)
					}
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	<-ch
}
