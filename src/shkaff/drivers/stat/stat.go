package stat

import (
	"fmt"
	"log"
	"shkaff/structs"

	"shkaff/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

const (
	URI_TEMPLATE = "tcp://%s:%d?debug=False"
)

type StatDB struct {
	uri string
	DB  *sqlx.DB
}

func InitStat() (s *StatDB) {
	cfg := config.InitControlConfig()
	var err error
	s = new(StatDB)
	s.uri = fmt.Sprintf(URI_TEMPLATE, cfg.STATBASE_HOST, cfg.STATBASE_PORT)
	if s.DB, err = sqlx.Connect("clickhouse", s.uri); err != nil {
		log.Fatalln(err)
	}
	return
}

func (s *StatDB) Insert(statMessage structs.StatMessage) (err error) {
	tx := s.DB.MustBegin()
	_, err = tx.NamedExec("INSERT INTO shkaff_stat (UserId, DbID, TaskId, NewOperator, SuccessOperator, FailOperator, ErrorOperator, NewDump, SuccessDump, FailDump, ErrorDump, NewRestore, SuccessRestore, FailRestore, ErrorRestore, CreateDate) VALUES (:UserId, :DbID, :TaskId, :NewOperator, :SuccessOperator, :FailOperator, :ErrorOperator, :NewDump, :SuccessDump, :FailDump, :ErrorDump, :NewRestore, :SuccessRestore, :FailRestore, :ErrorRestore, :CreateDate)", &statMessage)
	if err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		return
	}
	return
}
