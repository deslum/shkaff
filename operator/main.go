package main

import (
	"fmt"
	"log"
	"time"

	"encoding/json"
	"io/ioutil"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

const (
	CONFIG_FILE                   = "operator.json"
	DEFAULT_HOST                  = "localhost"
	DEFAULT_DATABASE_PORT         = 5432
	DEFAULT_AMQP_PORT             = 5672
	DEFAULT_DATABASE_DB           = "postgres"
	DEFAULT_REFRESH_DATABASE_SCAN = 60

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

	REQUEST_GET_STARTTIME = "SELECT task_id, start_time, verb, thread_count, ipv6, gzip, host, port, databases, db_user, db_password  FROM shkaff.tasks t INNER JOIN shkaff.db_settings db ON t.db_settings_id = db.db_id WHERE t.start_time <= to_timestamp(%d) AND t.is_active = true;"
	REQUESR_UPDATE_ACTIVE = "UPDATE shkaff.tasks SET is_active = $1 WHERE task_id = $2;"
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
	Verb        int       `json:"verb" db:"verb"`
	ThreadCount int       `json:"thread_count" db:"thread_count"`
	Gzip        bool      `json:"gzip" db:"gzip"`
	Ipv6        bool      `json:"ipv6" db:"ipv6"`
	Host        string    `json:"host" db:"host"`
	Port        int       `json:"port" db:"port"`
	StartTime   time.Time `json:"start_time" db:"start_time"`
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
	if cc.REFRESH_DATABASE_SCAN == 0 {
		cc.REFRESH_DATABASE_SCAN = DEFAULT_REFRESH_DATABASE_SCAN
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
	ps.refreshTimeScan = cf.REFRESH_DATABASE_SCAN
	return
}
func (ps *pSQL) Connect() (db *sqlx.DB) {
	var err error
	log.Println(ps.uri)
	db, err = sqlx.Connect(POSTGRES, ps.uri)
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

func isDublicateTask(opc []Task, task Task) (result bool) {
	for _, oc := range opc {
		if oc.TaskID == task.TaskID {
			return true
		}
	}
	return false
}

func remove(slice []Task, s int) []Task {
	return append(slice[:s], slice[s+1:]...)
}

func TaskSender(db *sqlx.DB, rmqChannel *amqp.Channel) {
	databases := make(map[string][]string)
	for {
		for numEl, cache := range opCache {
			queue, err := rmqChannel.QueueDeclare(
				"mongodb", // name
				true,      // durable
				false,     // delete when unused
				false,     // exclusive
				false,     // no-wait
				nil,       // arguments
			)
			if err != nil {
				log.Println("Queue", err)
			}
			if time.Now().Unix() > cache.StartTime.Unix() {
				json.Unmarshal([]byte(cache.Databases), &databases)
				if err != nil {
					log.Println("Unmarshal databases", err)
					continue
				}
				for base, sheets := range databases {
					for _, sheet := range sheets {
						cache.Database = base
						cache.Sheet = sheet
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
							continue
						} else {
							_, err = db.Exec(REQUESR_UPDATE_ACTIVE, false, cache.TaskID)
							if err != nil {
								log.Fatalln(err)
							}
						}
					}
				}
			}
			opCache = remove(opCache, numEl)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func Aggregator(db *sqlx.DB, refreshTimeScan int) {
	var task = Task{}
	var psqlUpdateTime *time.Timer
	psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
	for {
		select {
		case <-psqlUpdateTime.C:
			request := fmt.Sprintf(REQUEST_GET_STARTTIME, time.Now().Unix())
			rows, err := db.Queryx(request)
			if err != nil {
				log.Println(err)
				psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
				continue
			}
			for rows.Next() {
				if err := rows.StructScan(&task); err != nil {
					log.Println("Scan", err)
					psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
					continue
				}
				if !isDublicateTask(opCache, task) {
					opCache = append(opCache, task)
				}
			}
			psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
		}
	}
}

func main() {
	ch := make(chan bool)

	controlConfig := initControlConfig(CONFIG_FILE)
	controlConfig.validateConfig()

	pSQL := initPSQL(controlConfig)
	qp := initAMQP(controlConfig)

	db := pSQL.Connect()
	rmqChannel := qp.Connect()

	go Aggregator(db, controlConfig.REFRESH_DATABASE_SCAN)
	go TaskSender(db, rmqChannel)
	<-ch
}
