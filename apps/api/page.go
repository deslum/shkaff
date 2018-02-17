package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) General(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func (api *API) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{})
}

func (api *API) Databases(c *gin.Context) {
	c.HTML(http.StatusOK, "databases.tmpl", gin.H{})
}

func (api *API) ActiveTasks(c *gin.Context) {
	c.HTML(http.StatusOK, "tasks.tmpl", gin.H{
		"message": "active",
	})
}

func (api *API) UnactiveTasks(c *gin.Context) {
	c.HTML(http.StatusOK, "tasks.tmpl", gin.H{
		"message": "unactive",
	})
}
