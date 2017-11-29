package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"shkaff/drivers/maindb"
	"shkaff/drivers/mongodb"
	"shkaff/drivers/rmq/consumer"
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
	dumpChan     chan string
	databaseName string
	postgres     *maindb.PSQL
	workRabbit   *consumer.RMQ
}

func InitWorker() (ws *workersStarter) {
	ws = new(workersStarter)
	ws.workerWG = sync.WaitGroup{}
	dChan := make(chan string, 100)
	sChan := make(chan structs.StatMessage, 1000)
	startStatSender(sChan)
	worker := &worker{
		databaseName: "mongodb",
		dumpChan:     dChan,
		statChan:     sChan,
		postgres:     maindb.InitPSQL(),
		workRabbit:   consumer.InitAMQPConsumer(),
	}
	ws.workers = append(ws.workers, worker)
	return
}

func (ws *workersStarter) Run() {
	ws.workerWG.Add(1)
	for _, w := range ws.workers {
		go w.worker()
	}
	ws.workerWG.Wait()
}

func (w *worker) worker() {
	var task *structs.Task
	w.workRabbit.InitConnection("mongodb")
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
		log.Println(task.Database, task.Sheet, "3")
		dumpMsg, err := dbDriver.Dump(task)
		if err != nil {
			w.sendStatMessage(2, task.UserID, task.DBID, task.TaskID, err)
			log.Println("Worker", "Fail Dump message", err)
			continue
		}
		w.dumpChan <- dumpMsg
		w.sendStatMessage(3, task.UserID, task.DBID, task.TaskID, nil)
		message.Ack(true)
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
	var statMessage structs.StatMessage
	statMessage.Act = action
	statMessage.UserID = userID
	statMessage.DBID = dbid
	statMessage.TaskID = taskID
	statMessage.Error = err
	w.statChan <- statMessage
}
