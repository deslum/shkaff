package worker

import (
	"encoding/json"
	"log"
	"shkaff/config"
	"shkaff/drivers/maindb"
	"shkaff/drivers/mongodb"
	"shkaff/drivers/rmq/consumer"
	"shkaff/drivers/rmq/producer"
	"sync"
	"time"
)

var workerWG sync.WaitGroup = sync.WaitGroup{}

type Worker struct {
	postgres   *maindb.PSQL
	statRabbit *producer.RMQ
	workRabbit *consumer.RMQ
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

func (w *Worker) StartWorker() {
	var task Task
	log.Println("Start Worker")
	for message := range w.workRabbit.Msgs {
		if err := json.Unmarshal(message.Body, &task); err != nil {
			log.Println(err, "Failed JSON parse")
		}
		mg := mongodb.New(task.Host, task.Port, task.DBUser, task.DBPassword, task.Ipv6, task.Database, task.Sheet, task.Gzip, task.ThreadCount)
		mg.Dump()
		message.Ack(false)
	}
	workerWG.Done()
}

func InitWorker(cfg config.ShkaffConfig) (w *Worker) {
	w = &Worker{
		postgres:   maindb.InitPSQL(cfg),
		statRabbit: producer.InitAMQPProducer(cfg),
		workRabbit: consumer.InitAMQPConsumer(cfg),
	}
	return
}

func (w *Worker) Run() {
	workerWG.Add(1)
	go w.StartWorker()
	workerWG.Wait()
}
