package maindb

import (
	"fmt"
	"log"
	"shkaff/config"
	"shkaff/consts"

	"github.com/jmoiron/sqlx"
)

type PSQL struct {
	uri             string
	DB              *sqlx.DB
	RefreshTimeScan int
}

func InitPSQL(cfg config.ShkaffConfig) (ps *PSQL) {
	var err error
	ps = new(PSQL)
	ps.uri = fmt.Sprintf(consts.PSQL_URI_TEMPLATE, cfg.DATABASE_USER,
		cfg.DATABASE_PASS,
		cfg.DATABASE_HOST,
		cfg.DATABASE_PORT,
		cfg.DATABASE_DB)
	ps.RefreshTimeScan = cfg.REFRESH_DATABASE_SCAN
	if ps.DB, err = sqlx.Connect("postgres", ps.uri); err != nil {
		log.Fatalln(err)
	}
	return
}
