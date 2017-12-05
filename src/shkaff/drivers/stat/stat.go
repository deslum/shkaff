package stat

import (
	"fmt"
	"log"

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
