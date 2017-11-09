package operator

import (
	"fmt"
	"log"
	"shkaff/config"
	"shkaff/consts"
	"shkaff/drivers/maindb"
	"shkaff/drivers/rmq/producer"
	"time"
	"sync"

	"encoding/json"

	_ "github.com/lib/pq"
)

var (
	opCache []Task
	operatorWG sync.WaitGroup = sync.WaitGroup{}
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

func (oper *Operator) TaskSender() {
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

func (oper *Operator) Aggregator() {
	var task = Task{}
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

func InitOperator(cfg config.ShkaffConfig) (oper *Operator) {
	oper = &Operator{
		postgres: maindb.InitPSQL(cfg),
		rabbit:   producer.InitAMQPProducer(cfg),
	}
	return
}

func (oper *Operator) Run() {
	operatorWG.Add(1)
	log.Println("Start Operator")
	go oper.Aggregator()
	go oper.TaskSender()
	operatorWG.Wait()
}
