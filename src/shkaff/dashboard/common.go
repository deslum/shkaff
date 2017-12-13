package dashboard

import (
	"errors"
	"fmt"
	"shkaff/structs"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (api *API) checkUpdateParameters(c *gin.Context) (task structs.Task, err error) {
	var taskUpdate map[string]string
	var errStr string
	taskID := c.Param("TaskID")
	taskIDInt, err := strconv.Atoi(taskID)
	if err != nil {
		return
	}
	task, err = api.sql.GetTask(taskIDInt)
	if err != nil {
		return
	}
	c.BindJSON(&taskUpdate)
	for key, val := range taskUpdate {
		switch key {
		case "task_name":
			if val == "" {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.TaskName = val
		case "host":
			if val == "" {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.Host = val
		case "port":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt < 1024 && valInt > 65565 {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.Port = valInt
		case "verb":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt > 6 {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.Verb = valInt
		case "thread_count":
			valInt, err := strconv.Atoi(val)
			if err != nil || valInt > 10 {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.ThreadCount = valInt
		case "gzip":
			valBool, err := strconv.ParseBool(val)
			if err != nil {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.Gzip = valBool
		case "ipv6":
			valBool, err := strconv.ParseBool(val)
			if err != nil {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.Ipv6 = valBool
		case "start_time":
			layout := "2006-01-02T15:04:05.000Z"
			tm, err := time.Parse(layout, val)
			if err != nil {
				errStr = fmt.Sprintf("Bad %s %s", key, val)
				return task, errors.New(errStr)
			}
			task.StartTime = tm
		case "db_user":
			task.DBUser = val
		case "db_password":
			task.DBPassword = val
		case "database":
			task.Sheet = val
		case "sheet":
			task.Database = val
		}
	}
	return
}
