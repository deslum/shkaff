package operator

import (
	"encoding/json"
	"fmt"
	"log"
	"shkaff/consts"
	"shkaff/drivers/maindb"
	"shkaff/drivers/mongodb"
	"shkaff/drivers/rmq/producer"
	"shkaff/structs"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type operator struct {
	tasksChan  chan structs.Task
	operatorWG sync.WaitGroup
	postgres   *maindb.PSQL
	rabbit     *producer.RMQ
}

func InitOperator() (oper *operator) {
	oper = &operator{
		postgres:  maindb.InitPSQL(),
		rabbit:    producer.InitAMQPProducer("mongodb"),
		tasksChan: make(chan structs.Task),
	}
	return
}

func (oper *operator) Run() {
	oper.operatorWG = sync.WaitGroup{}
	oper.operatorWG.Add(2)
	log.Println("Start Operator")
	go oper.aggregator()
	go oper.taskSender()
	oper.operatorWG.Wait()
}

func (oper *operator) taskSender() {
	var messages []structs.Task
	rabbit := oper.rabbit
	for task := range oper.tasksChan {
		switch dbType := task.DBType; dbType {
		case "mongodb":
			messages = mongodb.GetMessages(task)
		default:
			log.Printf("Driver for Database %s not found", task.DBType)
			continue
		}
		for _, msg := range messages {
			body, err := json.Marshal(msg)
			if err != nil {
				log.Println("Body", err)
				continue
			}
			if err := rabbit.Publish(body); err != nil {
				log.Println("Publish", err)
				continue
			}
		}
	}
}

func (oper *operator) aggregator() {
	var task = structs.Task{}
	db := oper.postgres.DB
	refreshTimeScan := oper.postgres.RefreshTimeScan
	psqlUpdateTime := time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
	for {
		select {
		case <-psqlUpdateTime.C:
			tsNow := time.Now()
			request := fmt.Sprintf(consts.REQUEST_GET_STARTTIME, tsNow.Month, tsNow.Day, tsNow.Hour, tsNow.Minute)
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
				oper.tasksChan <- task
			}
			psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
		}
	}
}
