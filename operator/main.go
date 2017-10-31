package main

import (
	"fmt"
	"log"
	"time"

	"database/sql"
	"encoding/json"
	"io/ioutil"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
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

	REQUEST_GET_STARTTIME = "SELECT * FROM shkaff.tasks WHERE start_time <= to_timestamp(%d) AND is_active = 1"
	REQUESR_UPDATE_ACTIVE = "UPDATE shkaff.tasks SET is_active = $1 WHERE task_id = $2;"

	REFRESH_DATABASE_SCAN = 5
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

type Task struct {
	Taskid           int       `json:"taskid"`
	Task_name        string    `json:"task_name"`
	Verbose          int       `json:"verbose"`
	Start_time       time.Time `json:"start_time"`
	Is_active        bool      `json:"is_active"`
	Thread_count     int       `json:"thread_count"`
	Db_settings_id   int64     `json:"db_settings_id"`
	Db_settings_type int64     `json:"db_settings_type"`
}

var opCache []Task

func isDublicateTask(opc []Task, task Task) (result bool) {
	for _, oc := range opc {
		if oc.Taskid == task.Taskid {
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
				request := fmt.Sprintf(REQUEST_GET_STARTTIME, time.Now().Unix())
				if rows, err := db.Query(request); err != nil {
					log.Println(err)
				} else {
					for rows.Next() {
						if err := rows.Scan(&task.Taskid,
							&task.Task_name,
							&task.Verbose,
							&task.Start_time,
							&task.Is_active,
							&task.Thread_count,
							&task.Db_settings_id,
							&task.Db_settings_type); err != nil {
							log.Println("Scan", err)
						}
						if !isDublicateTask(opCache, task) {
							opCache = append(opCache, task)
						}
					}
					if err = rows.Err(); err != nil {
						log.Println("Rows", err)
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
					log.Println("Queue", err)
				}
				if time.Now().Unix() > 0 {
					body, err := json.Marshal(cache)
					if err != nil {
						log.Println("Body", err)
						continue
					}
					pub := amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					}
					if err := rmqChannel.Publish("", queue.Name, false, false, pub); err != nil {
						log.Println("Publish", err)
					} else {
						_, err = db.Exec(REQUESR_UPDATE_ACTIVE, 0, cache.Taskid)
						if err != nil {
							log.Fatalln(err)
						}
						opCache = remove(opCache, numEl)
					}
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	<-ch
}
