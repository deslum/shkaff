package operator

import (
	"fmt"
	"log"
	"shkaff/consts"
	"shkaff/drivers/maindb"
	"shkaff/drivers/rmq/producer"
	"shkaff/structs"
	"sync"
	"time"

	"encoding/json"

	_ "github.com/lib/pq"
)

var (
	opCache    []*structs.Task
	operatorWG sync.WaitGroup = sync.WaitGroup{}
)

type operator struct {
	postgres *maindb.PSQL
	rabbit   *producer.RMQ
}

func InitOperator() (oper *operator) {
	oper = &operator{
		postgres: maindb.InitPSQL(),
		rabbit:   producer.InitAMQPProducer("mongodb"),
	}
	return
}

func (oper *operator) Run() {
	operatorWG.Add(2)
	log.Println("Start Operator")
	go oper.aggregator()
	go oper.taskSender()
	operatorWG.Wait()
}

func isDublicateTask(opc []*structs.Task, task *structs.Task) (result bool) {
	for _, oc := range opc {
		if oc.TaskID == task.TaskID {
			return true
		}
	}
	return false
}

func remove(slice []*structs.Task, s int) []*structs.Task {
	return append(slice[:s], slice[s+1:]...)
}

func (oper *operator) taskSender() {
	databases := make(map[string][]string)
	db := oper.postgres.DB
	rabbit := oper.rabbit
	for {
		for numEl, cache := range opCache {
			if time.Now().Unix() > cache.StartTime.Unix() {
				if err := json.Unmarshal([]byte(cache.Databases), &databases); err != nil {
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
						if err := rabbit.Publish(body); err != nil {
							log.Println("Publish", err)
							continue
						}
						if _, err = db.Exec(consts.REQUESR_UPDATE_ACTIVE, false, cache.TaskID); err != nil {
							log.Fatalln(err)
							continue
						}
					}
				}
			}
			opCache = remove(opCache, numEl)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (oper *operator) aggregator() {
	var task = &structs.Task{}
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
				if !isDublicateTask(opCache, task) {
					opCache = append(opCache, task)
				}
			}
			psqlUpdateTime = time.NewTimer(time.Duration(refreshTimeScan) * time.Second)
		}
	}
}
