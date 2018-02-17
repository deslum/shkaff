package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (api *API) getUsers(c *gin.Context) {
	users, err := api.psql.GetUsers()
	if err != nil {
		api.log.Error(err)
		c.JSON(400, "")
		return
	}
	c.JSON(http.StatusOK, users)
	return
}

func (api *API) getDatabases(c *gin.Context) {
	dbTypes := []string{"mongodb", "postgresql", "mysql"}
	isActive := c.Query("is_active")
	_, err := strconv.ParseBool(isActive)
	if err != nil {
		api.log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not valid active value"})
		return
	}
	dbType := c.Query("type")
	if res := isExistStringArr(dbTypes, dbType); !res {
		api.log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not found db type"})
		return
	}
	databases, err := api.psql.GetDatabases(isActive, dbType)
	if err != nil {
		api.log.Error(err)
		c.JSON(400, "")
		return
	}
	c.JSON(http.StatusOK, databases)
	return
}

func (api *API) getTasks(c *gin.Context) {
	isActive := c.Query("is_active")
	_, err := strconv.ParseBool(isActive)
	if err != nil {
		api.log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not valid active value"})
		return
	}
	tasks, err := api.psql.GetTasks(isActive)
	if err != nil {
		api.log.Error(err)
		c.JSON(400, "")
		return
	}
	c.JSON(http.StatusOK, tasks)
	return
}

func (api *API) changeTaskStatus(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		api.log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not valid TaskID"})
		return
	}
	activate := c.Query("activate")
	status, err := strconv.ParseBool(activate)
	if err != nil {
		api.log.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"Error": "Not valid activate value"})
		return
	}
	err = api.psql.ChangeTaskStatus(taskID, status)
	if err != nil {
		api.log.Error(err)
		c.JSON(400, "")
		return
	}
	c.JSON(http.StatusOK, "")
	return
}

func isExistStringArr(s []string, v string) (result bool) {
	for _, k := range s {
		if k == v {
			return true
		}
	}
	return false
}
