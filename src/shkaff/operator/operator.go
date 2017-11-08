package operator

import (
	"fmt"
	"log"
	"shkaff/config"
	"shkaff/drivers/maindb"
	"shkaff/drivers/rmq/producer"
	"time"

	"encoding/json"

	_ "github.com/bmizerany/pq"
	"github.com/jmoiron/sqlx"
	"github.com/streadway/amqp"
)

const (
	REQUEST_GET_STARTTIME = "SELECT task_id, start_time, verb, thread_count, ipv6, gzip, host, port, databases, db_user, db_password,tp.\"type\" as db_type FROM shkaff.tasks t INNER JOIN shkaff.db_settings db ON t.db_settings_id = db.db_id INNER JOIN shkaff.types tp ON tp.type_id = db.\"type\" WHERE t.start_time <= to_timestamp(%d) AND t.is_active = true;"
	REQUESR_UPDATE_ACTIVE = "UPDATE shkaff.tasks SET is_active = $1 WHERE task_id = $2;"
)

var (
	opCache []Task
)

type Operator struct {
	postgres *maindb.PSQL
	rabbit   *producer.RMQ
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
	StartTime   time.Time `json:"start_time" db:"start_time"`
	DBUser      string    `json:"db_user" db:"db_user"`
	DBPassword  string    `json:"db_password" db:"db_password"`
	Database    string    `json:"database"`
	Sheet       string    `json:"sheet"`
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
				cache.DBType, // name
				true,         // durable
				false,        // delete when unused
				false,        // exclusive
				false,        // no-wait
				nil,          // arguments
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

func InitOperator(cfg config.ShkaffConfig) (o *Operator) {
	o = &Operator{
		postgres: maindb.InitPSQL(cfg),
		rabbit:   producer.InitAMQPProducer(cfg),
	}
	return
}

func (o *Operator) Run() {
	ch := make(chan bool)
	go Aggregator(o.postgres.DB, o.postgres.RefreshTimeScan)
	go TaskSender(o.postgres.DB, o.rabbit.Channel)
	<-ch
}
