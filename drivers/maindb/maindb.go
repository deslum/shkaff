package maindb

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"fmt"
	"shkaff/internal/consts"
	"shkaff/internal/options"
	"shkaff/internal/structs"

	"github.com/jmoiron/sqlx"
)

type PSQL struct {
	uri             string
	DB              *sqlx.DB
	RefreshTimeScan int
}

func InitPSQL() (ps *PSQL) {
	var err error
	cfg := options.InitControlConfig()
	ps = new(PSQL)
	ps.uri = fmt.Sprintf(consts.PSQL_URI_TEMPLATE, cfg.DATABASE_USER,
		cfg.DATABASE_PASS,
		cfg.DATABASE_HOST,
		cfg.DATABASE_PORT,
		cfg.DATABASE_DB)
	ps.RefreshTimeScan = cfg.REFRESH_DATABASE_SCAN
	for {
		ps.DB, err = sqlx.Connect("postgres", ps.uri)
		if err == nil {
			break
		}
		log.Printf("PSQL: %s not connected. Error %s\n", ps.uri, err.Error())
		time.Sleep(time.Second * 5)
	}
	return
}

func (ps *PSQL) GetTask(taskId int, isSimple bool) (task structs.APITask, err error) {
	var requestString string
	if isSimple {
		requestString = `SELECT * FROM shkaff.tasks WHERE task_id = $1 AND is_delete = false`
	} else {
		requestString = `SELECT 
		task_id,
		task_name,
		is_active,
		db_id,
		databases,
		"verb",
		thread_count,
		gzip,
		ipv6,
		array_to_string(months, ',', '') as months,
		array_to_string(days, ',', '') as days,
		array_to_string(hours, ',', '') as hours,
		minutes 
	FROM shkaff.tasks 
    WHERE task_id = $1 and is_delete = false`
	}
	err = ps.DB.Get(&task, requestString, taskId)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) GetLastTaskID() (id int, err error) {
	err = ps.DB.Get(id, "SELECT Count(*) FROM shkaff.tasks WHERE is_delete = false")
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) GetTaskByName(taskName string) (task structs.APITask, err error) {
	requestString := `SELECT 
		task_id,
		task_name,
		is_active,
		db_id,
		databases,
		"verb",
		thread_count,
		gzip,
		ipv6,
		array_to_string(months, ',', '') as months,
		array_to_string(days, ',', '') as days,
		array_to_string(hours, ',', '') as hours,
		minutes 
	FROM shkaff.tasks 
    WHERE task_name = $1 AND is_delete = false`
	err = ps.DB.Get(&task, requestString, taskName)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) CreateTask(setStrings map[string]interface{}) (result sql.Result, err error) {
	var keys, dottedKeys []string
	var returnID int
	for key, value := range setStrings {
		switch key {
		case "db_id":
			err = ps.DB.Get(&returnID, `SELECT user_id FROM shkaff.db_settings WHERE db_id = $1 AND is_delete = false`, value.(int))
			if err != nil {
				errStr := fmt.Sprintf("Databases with ID %d not found", value.(int))
				return nil, errors.New(errStr)
			}
		}
		keys = append(keys, key)
		dottedKeys = append(dottedKeys, ":"+key)
	}
	cols := strings.Join(keys, ",")
	dottedCols := strings.Join(dottedKeys, ",")
	sqlString := fmt.Sprintf("INSERT INTO shkaff.tasks (%s) VALUES (%s)", cols, dottedCols)
	result, err = ps.DB.NamedExec(sqlString, setStrings)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) UpdateTask(taskIDInt int, setStrings map[string]interface{}) (result sql.Result, err error) {
	var keys []string
	var returnID int
	for key, value := range setStrings {
		switch key {
		case "db_id":
			err = ps.DB.Get(&returnID, `SELECT user_id FROM shkaff.db_settings WHERE db_id = $1 AND is_delete = false`, value.(int))
			if err != nil {
				errStr := fmt.Sprintf("Databases with ID %d not found", value.(int))
				return nil, errors.New(errStr)
			}
		}
		keys = append(keys, fmt.Sprintf("%s=:%s", key, key))
	}
	cols := strings.Join(keys, ",")
	sqlString := fmt.Sprintf("UPDATE shkaff.tasks SET %s WHERE task_id = %d", cols, taskIDInt)
	log.Println(sqlString)
	result, err = ps.DB.NamedExec(sqlString, setStrings)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func (ps *PSQL) DeleteTask(taskId int) (result sql.Result, err error) {
	result, err = ps.DB.Exec("UPDATE shkaff.tasks SET is_delete = true WHERE task_id = $1", taskId)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) GetDatabase(databaseId int) (database structs.APIDatabase, err error) {
	requestString := `SELECT * FROM shkaff.db_settings WHERE db_id = $1 and is_delete = false`
	err = ps.DB.Get(&database, requestString, databaseId)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) UpdateDatabase(databaseIDInt int, setStrings map[string]interface{}) (result sql.Result, err error) {
	var keys []string
	var returnID int
	for key, value := range setStrings {
		switch key {
		case "user_id":
			err = ps.DB.Get(&returnID, `SELECT user_id FROM shkaff.users WHERE user_id = $1 AND is_active = true and is_delete = false`, value.(int))
			if err != nil {
				errStr := fmt.Sprintf("Active user with ID %d not found", value.(int))
				return nil, errors.New(errStr)
			}
		case "type_id":
			err = ps.DB.Get(&returnID, `SELECT type_id FROM shkaff.types WHERE type_id = $1 and is_delete = false`, value.(int))
			if err != nil {
				errStr := fmt.Sprintf("Databases with typeID %d not found", value.(int))
				return nil, errors.New(errStr)
			}
		}
		keys = append(keys, fmt.Sprintf("%s=:%s", key, key))
	}
	cols := strings.Join(keys, ",")
	sqlString := fmt.Sprintf("UPDATE shkaff.db_settings SET %s WHERE db_id = %d and is_delete = false", cols, databaseIDInt)
	result, err = ps.DB.NamedExec(sqlString, setStrings)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) CreateDatabase(setStrings map[string]interface{}) (result sql.Result, err error) {
	var keys, dottedKeys []string
	var returnID int
	for key, value := range setStrings {
		switch key {
		case "user_id":
			err = ps.DB.Get(&returnID, `SELECT user_id FROM shkaff.users WHERE user_id = $1 AND is_active = true and is_delete = false`, value.(int))
			if err != nil {
				errStr := fmt.Sprintf("Active user with ID %d not found", value.(int))
				return nil, errors.New(errStr)
			}
		case "type_id":
			err = ps.DB.Get(&returnID, `SELECT type_id FROM shkaff.types WHERE type_id = $1`, value.(int))
			if err != nil {
				errStr := fmt.Sprintf("Databases with typeID %d not found", value.(int))
				return nil, errors.New(errStr)
			}
		}
		keys = append(keys, key)
		dottedKeys = append(dottedKeys, ":"+key)
	}
	cols := strings.Join(keys, ",")
	dottedCols := strings.Join(dottedKeys, ",")
	sqlString := fmt.Sprintf("INSERT INTO shkaff.db_settings (%s) VALUES (%s)", cols, dottedCols)
	log.Println(sqlString)
	result, err = ps.DB.NamedExec(sqlString, setStrings)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) DeleteDatabase(databaseID int) (result sql.Result, err error) {
	result, err = ps.DB.Exec("UPDATE shkaff.db_settings SET is_delete = true WHERE db_id = $1", databaseID)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) GetUser(userId int) (user structs.APIUser, err error) {
	requestString := `SELECT * FROM shkaff.users WHERE user_id = $1 AND is_delete = false`
	err = ps.DB.Get(&user, requestString, userId)
	if err != nil {
		return
	}
	return
}
func (ps *PSQL) GetUserByToken(token string) (isExist bool, err error) {
	var t string
	requestString := `SELECT user_id FROM shkaff.users WHERE api_token = $1 AND is_delete = false`
	err = ps.DB.Get(&t, requestString, token)
	if err != nil {
		return false, err
	}
	if t != "" {
		return true, err
	}
	return false, err
}

func (ps *PSQL) UpdateUser(userIDInt int, setStrings map[string]interface{}) (result sql.Result, err error) {
	var keys []string
	for key := range setStrings {
		keys = append(keys, fmt.Sprintf("%s=:%s", key, key))
	}
	cols := strings.Join(keys, ",")
	sqlString := fmt.Sprintf("UPDATE shkaff.users SET %s WHERE user_id = %d AND is_delete = false", cols, userIDInt)
	result, err = ps.DB.NamedExec(sqlString, setStrings)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) CreateUser(setStrings map[string]interface{}) (result sql.Result, err error) {
	var keys, dottedKeys []string
	for key := range setStrings {
		keys = append(keys, key)
		dottedKeys = append(dottedKeys, ":"+key)
	}
	cols := strings.Join(keys, ",")
	dottedCols := strings.Join(dottedKeys, ",")
	sqlString := fmt.Sprintf("INSERT INTO shkaff.users (%s) VALUES (%s)", cols, dottedCols)
	result, err = ps.DB.NamedExec(sqlString, setStrings)
	if err != nil {
		return
	}
	return
}

func (ps *PSQL) DeleteUser(userID int) (result sql.Result, err error) {
	result, err = ps.DB.Exec("UPDATE shkaff.users SET is_delete = true WHERE user_id = $1", userID)
	if err != nil {
		return
	}
	return
}
