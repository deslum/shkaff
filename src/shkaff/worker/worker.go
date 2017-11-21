package worker

import (
	"encoding/json"
	"errors"
	"fmt"
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

type workersStarter struct {
	workerWG sync.WaitGroup
	workers  []*worker
}

type worker struct {
	statChan     chan structs.StatMessage
	databaseName string
	postgres     *maindb.PSQL
	statRabbit   *producer.RMQ
	workRabbit   *consumer.RMQ
}

func InitWorker() (ws *workersStarter) {
	cfg := config.InitControlConfig()
	ws = new(workersStarter)
	for database, workerCount := range cfg.WORKERS {
		for count := 0; count < workerCount; count++ {
			worker := &worker{
				databaseName: database,
				statChan:     make(chan structs.StatMessage, workerCount),
				postgres:     maindb.InitPSQL(),
				statRabbit:   producer.InitAMQPProducer("shkaff_stat"),
				workRabbit:   consumer.InitAMQPConsumer(),
			}
			ws.workers = append(ws.workers, worker)
		}
	}
	return
}

func (ws *workersStarter) Run() {
	var workerWGCount = 0
	ws.workerWG = sync.WaitGroup{}
	for _, w := range ws.workers {
		workerWGCount += 2
		ws.workerWG.Add(workerWGCount)
		go w.worker()
		go w.statSender()
	}
	ws.workerWG.Wait()
}

func (w *worker) statSender() {
	for {
		statMsg, ok := <-w.statChan
		if !ok {
			break
		}
		msg, err := json.Marshal(statMsg)
		if err != nil {
			log.Println(err)
			continue
		}
		err = w.statRabbit.Publish(msg)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func (w *worker) worker() {
	var task *structs.Task
	w.workRabbit.InitConnection(w.databaseName)
	dbDriver, err := w.getDatabaseType()
	if err != nil {
		log.Println("Worker", err)
		return
	}
	log.Printf("Start Worker for %s\n", w.databaseName)
	for message := range w.workRabbit.Msgs {
		err := json.Unmarshal(message.Body, &task)
		w.sendStatMessage(0, task.UserID, task.DBID, task.TaskID, nil)
		if err != nil {
			w.sendStatMessage(2, task.UserID, task.DBID, task.TaskID, err)
			log.Println("Worker", err, "Failed JSON parse")
			continue
		}
		dbDriver.Dump(task)
		err = message.Ack(false)
		if err != nil {
			w.sendStatMessage(2, task.UserID, task.DBID, task.TaskID, err)
			log.Println("Worker", "Fail Ack message", err)
			continue
		}
		w.sendStatMessage(3, task.UserID, task.DBID, task.TaskID, nil)
	}
}

func (w *worker) getDatabaseType() (dbDriver databases.DatabaseDriver, err error) {
	switch w.databaseName {
	case "mongodb":
		dbDriver = mongodb.InitDriver()
		return dbDriver, nil
	default:
		answer := fmt.Sprintf("Driver %s not found", w.databaseName)
		return nil, errors.New(answer)
	}
}

func (w *worker) sendStatMessage(action structs.Action, userID, dbid, taskID int, err error) {
	var statMessage = new(structs.StatMessage)
	statMessage.Act = action
	statMessage.UserID = userID
	statMessage.DBID = dbid
	statMessage.TaskID = taskID
	statMessage.Error = err
	w.statChan <- *statMessage
}
