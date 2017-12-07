package stat

import (
	"database/sql"
	"fmt"
	"log"
	"shkaff/config"
	"shkaff/structs"
	"sync"
	"time"

	_ "github.com/kshvakov/clickhouse"
)

const (
	URI_TEMPLATE   = "tcp://%s:%d?debug=False"
	CHECKOUT_TIME  = 15
	INSERT_REQUEST = "INSERT INTO shkaff_stat (UserId, DbID, TaskId, NewOperator, SuccessOperator, FailOperator, ErrorOperator, NewDump, SuccessDump, FailDump, ErrorDump, NewRestore, SuccessRestore, FailRestore, ErrorRestore, CreateDate) VALUES (:UserId, :DbID, :TaskId, :NewOperator, :SuccessOperator, :FailOperator, :ErrorOperator, :NewDump, :SuccessDump, :FailDump, :ErrorDump, :NewRestore, :SuccessRestore, :FailRestore, :ErrorRestore, :CreateDate)"
)

type StatDB struct {
	mutex           sync.Mutex
	uri             string
	statMessageList []structs.StatMessage
	DB              *sql.DB
}

func InitStat() (s *StatDB) {
	cfg := config.InitControlConfig()
	var err error
	s = new(StatDB)
	s.mutex = sync.Mutex{}
	s.uri = fmt.Sprintf(URI_TEMPLATE, cfg.STATBASE_HOST, cfg.STATBASE_PORT)
	if s.DB, err = sql.Open("clickhouse", s.uri); err != nil {
		log.Fatalln(err)
	}
	go s.checkout()
	return
}

func (s *StatDB) Insert(statMessage structs.StatMessage) (err error) {
	s.mutex.Lock()
	s.statMessageList = append(s.statMessageList, statMessage)
	s.mutex.Unlock()
	return
}

func (s *StatDB) checkout() {
	for {
		timer := time.NewTimer(time.Second * CHECKOUT_TIME)
		select {
		case <-timer.C:
			if len(s.statMessageList) > 0 {
				s.inserBulk()
			}
		}
	}
}

func (s *StatDB) inserBulk() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	tx, err := s.DB.Begin()
	if err != nil {
		log.Println(err)
	}
	stmt, err := tx.Prepare(INSERT_REQUEST)
	if err != nil {
		log.Println(err)
	}
	for _, sm := range s.statMessageList {
		_, err = stmt.Exec(sm.UserID, sm.DbID, sm.TaskID, sm.NewOperator, sm.SuccessOperator,
			sm.FailOperator, sm.ErrorOperator, sm.NewDump, sm.SuccessDump, sm.FailDump,
			sm.ErrorDump, sm.NewRestore, sm.SuccessRestore, sm.FailRestore,
			sm.ErrorRestore, sm.CreateDate)
		if err != nil {
			log.Println(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}
	s.dropList()
}

func (s *StatDB) dropList() {
	s.statMessageList = []structs.StatMessage{}
}
