package operator

import (
	"encoding/json"
	"fmt"
	"shkaff/drivers/cache"
	"shkaff/drivers/maindb"
	"shkaff/drivers/mongodb"
	"shkaff/drivers/rmq/producer"
	"shkaff/internal/consts"
	"shkaff/internal/logger"
	"shkaff/internal/structs"
	"sync"
	"time"

	_ "github.com/lib/pq"
	logging "github.com/op/go-logging"
)

type Operator struct {
	tasksChan  chan structs.Task
	operatorWG sync.WaitGroup
	postgres   *maindb.PSQL
	rabbit     *producer.RMQ
	taskCache  *cache.Cache
	log        *logging.Logger
}

func InitOperator() (oper *Operator) {
	var err error
	oper = &Operator{
		postgres:  maindb.InitPSQL(),
		rabbit:    producer.InitAMQPProducer("mongodb"),
		tasksChan: make(chan structs.Task),
		log:       logger.GetLogs("Operator"),
	}
	oper.taskCache, err = cache.InitCacheDB()
	if err != nil {
		oper.log.Fatal(err)
	}
	return
}

func (oper *Operator) Run() {
	oper.operatorWG = sync.WaitGroup{}
	oper.operatorWG.Add(2)
	go oper.Aggregator()
	go oper.TaskSender()
	oper.log.Info("Start Operator")
	oper.operatorWG.Wait()
}

func (oper *Operator) Stop() {
	for i := 0; i < 2; i++ {
		oper.operatorWG.Done()
	}
	oper.log.Info("Stop Operator")
}

func (oper *Operator) TaskSender() {
	var messages []structs.Task
	rabbit := oper.rabbit
	for task := range oper.tasksChan {
		switch dbType := task.DBType; dbType {
		case "mongodb":
			messages = mongodb.GetMessages(task)
		default:
			err := fmt.Sprintf("Driver for Database %s not found", task.DBType)
			oper.log.Info(err)
			continue
		}
		for _, msg := range messages {
			body, err := json.Marshal(msg)
			if err != nil {
				oper.log.Error(err)
				continue
			}
			if err := rabbit.Publish(body); err != nil {
				oper.log.Error(err)
				continue
			}
		}
	}
}

func (oper *Operator) Aggregator() {
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
				oper.log.Error(err)
				psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
				continue
			}
			for rows.Next() {
				if err := rows.StructScan(&task); err != nil {
					oper.log.Error(err)
					psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
					continue
				}
				isExist, err := oper.taskCache.ExistKV(task.UserID, task.DBSettingsID, task.TaskID)
				if err != nil {
					oper.log.Error(err)
					psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
					continue
				}
				if !isExist {
					err := oper.taskCache.SetKV(task.UserID, task.DBSettingsID, task.TaskID)
					if err != nil {
						oper.log.Error(err)
						psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
						continue
					}
					oper.tasksChan <- task
				}

			}
			psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
		}
	}
}
