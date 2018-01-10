package maindb

import (
	"database/sql"

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

	err = ps.DB.Get(&task, `SELECT 
		task_id,
		task_name,
		is_active,
		db_id,
		databases,
		"verbose",
		thread_count,
		gzip,
		ipv6,
		array_to_string(months, ',', '') as months,
		array_to_string(days, ',', '') as days,
		array_to_string(hours, ',', '') as hours,
		minutes 
	FROM shkaff.tasks 
    WHERE task_id = $1`, taskId)
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

func (ps *PSQL) UpdateTask(taskIDInt int, setStrings string) (result sql.Result, err error) {
	sqlString := fmt.Sprintf("UPDATE shkaff.tasks SET %s WHERE task_id = %d", setStrings, taskIDInt)
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
