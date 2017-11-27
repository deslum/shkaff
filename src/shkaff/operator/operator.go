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
	opCache    []structs.Task
	operatorWG sync.WaitGroup
	postgres   *maindb.PSQL
	rabbit     *producer.RMQ
}

func InitOperator() (oper *operator) {
	oper = &operator{
		postgres: maindb.InitPSQL(),
		rabbit:   producer.InitAMQPProducer("mongodb"),
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

func isDublicateTask(opc []structs.Task, task structs.Task) (result bool) {
	for _, oc := range opc {
		if oc.TaskID == task.TaskID {
			return true
		}
	}
	return false
}

func remove(slice []structs.Task, s int) []structs.Task {
	return append(slice[:s], slice[s+1:]...)
}

func (oper *operator) taskSender() {
	var messages []structs.Task
	db := oper.postgres.DB
	rabbit := oper.rabbit
	for {
		for numEl, cache := range oper.opCache {
			if time.Now().Unix() < cache.StartTime.Unix() {
				continue
			}
			switch dbType := cache.DBType; dbType {
			case "mongodb":
				messages = mongodb.GetMessages(cache)
			default:
				log.Printf("Driver for Database %s not found", cache.DBType)
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
			if _, err := db.Exec(consts.REQUESR_UPDATE_ACTIVE, false, cache.TaskID); err != nil {
				log.Fatalln(err)
				continue
			}
			oper.opCache = remove(oper.opCache, numEl)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (oper *operator) aggregator() {
	var task = structs.Task{}
	var psqlUpdateTime *time.Timer
	db := oper.postgres.DB
	refreshTimeScan := oper.postgres.RefreshTimeScan
	psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
	for {
		select {
		case <-psqlUpdateTime.C:
			request := fmt.Sprintf(consts.REQUEST_GET_STARTTIME, time.Now().Unix())
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
				if !isDublicateTask(oper.opCache, task) {
					oper.opCache = append(oper.opCache, task)
				}
			}
			psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
		}
	}
}
