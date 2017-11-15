package worker

import (
	"encoding/json"
	"errors"
	"log"
	"shkaff/config"
	"shkaff/drivers/maindb"
	"shkaff/drivers/mongodb"
	"shkaff/drivers/rmq/consumer"
	"shkaff/drivers/rmq/producer"
	"shkaff/structs"
	"shkaff/structs/databases"
	"sync"
)

var workerWG sync.WaitGroup = sync.WaitGroup{}

type WorkersStarter struct {
	workers []*worker
}

type worker struct {
	databaseName string
	postgres     *maindb.PSQL
	statRabbit   *producer.RMQ
	workRabbit   *consumer.RMQ
}

func (w *worker) StartWorker() {
	var task *structs.Task
	w.workRabbit.InitConnection(w.databaseName)
	dbDriver, err := w.getDatabaseType(); 
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Start Worker for %s\n", w.databaseName)
	for message := range w.workRabbit.Msgs {
		if err := json.Unmarshal(message.Body, &task); err != nil {
			log.Println(err, "Failed JSON parse")
			return
		}
		dbDriver.Dump(task)
		message.Ack(false)
	}
	workerWG.Done()
}

func (w *worker) getDatabaseType() (dbDriver databases.DatabaseDriver, err error) {
	switch w.databaseName {
	case "mongodb":
		dbDriver = mongodb.InitDriver()
		return dbDriver, nil
	default:
		return nil, errors.New("Driver not found")
	}
}

func InitWorker(cfg config.ShkaffConfig) (ws *WorkersStarter) {
	ws = new(WorkersStarter)
	for database, workerCount := range cfg.WORKERS {
		for count := 0; count < workerCount; count++ {
			worker := &worker{
				databaseName: database,
				postgres:     maindb.InitPSQL(cfg),
				statRabbit:   producer.InitAMQPProducer(cfg),
				workRabbit:   consumer.InitAMQPConsumer(cfg),
			}
			ws.workers = append(ws.workers, worker)
		}
	}
	return
}

func (ws *WorkersStarter) Run() {
	for _, w := range ws.workers {
		workerWG.Add(1)
		go w.StartWorker()
	}
	workerWG.Wait()
}
