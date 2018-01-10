package maindb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"shkaff/config"
	"shkaff/consts"
	"shkaff/structs"

	"github.com/jmoiron/sqlx"
)

type PSQL struct {
	uri             string
	DB              *sqlx.DB
	RefreshTimeScan int
}

func InitPSQL() (ps *PSQL) {
	var err error
	cfg := config.InitControlConfig()
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

func (ps *PSQL) GetTask(taskId int) (task structs.APITask, err error) {
	var arr []uint8
	err = ps.DB.Get(&task, "SELECT * FROM shkaff.tasks WHERE task_id = $1", taskId)

	json.Unmarshal(task.Months, &arr)
	log.Println(arr)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) GetLastTaskID() (id int, err error) {
	err = ps.DB.Get(id, "SELECT Count(*) FROM shkaff.tasks")
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) UpdateTask(sqlString string) (result sql.Result, err error) {
	log.Println(sqlString)
	result, err = ps.DB.Exec(sqlString)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) DeleteTask(taskId int) (result sql.Result, err error) {
	result, err = ps.DB.Exec("DELETE FROM shkaff.tasks WHERE task_id = $1", taskId)
	if err != nil {
		return
	}
	return
}
