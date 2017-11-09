package maindb

import (
	"fmt"
	"log"
	"shkaff/config"

	"github.com/jmoiron/sqlx"
)

const (
	uriTemplate = "postgres://%s:%s@%s:%d/%s?sslmode=disable"
)

type PSQL struct {
	uri             string
	DB              *sqlx.DB
	RefreshTimeScan int
}

func InitPSQL(cfg config.ShkaffConfig) (ps *PSQL) {
	var err error
	ps = new(PSQL)
	ps.uri = fmt.Sprintf(uriTemplate, cfg.DATABASE_USER,
		cfg.DATABASE_PASS,
		cfg.DATABASE_HOST,
		cfg.DATABASE_PORT,
		cfg.DATABASE_DB)
	fmt.Println(ps.uri)
	ps.RefreshTimeScan = cfg.REFRESH_DATABASE_SCAN
	if ps.DB, err = sqlx.Connect("postgres", ps.uri); err != nil {
		log.Println(err)
	}
	return
}
