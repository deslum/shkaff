package dashboard

import (
	"fmt"
	"log"
	"net/http"
	"shkaff/config"

	"github.com/gin-gonic/gin"
)

type API struct {
	cfg    *config.ShkaffConfig
	router *gin.Engine
}

func InitAPI() (api *API) {
	api = new(API)
	api.cfg = config.InitControlConfig()
	api.router = gin.Default()
	v1 := api.router.Group("/api/v1")
	{
		v1.PUT("/CreateTask", api.createTask)
		v1.POST("/UpdateTask", api.updateTask)
		v1.GET("/GetTask", api.getTask)
		v1.DELETE("/DeleteTask", api.deleteTask)
	}

	return
}

func (api *API) Run() {
	log.Println("Start Dashboard")
	uri := fmt.Sprintf("%s:%d", api.cfg.SHKAFF_UI_HOST, api.cfg.SHKAFF_UI_PORT)
	err := api.router.Run(uri)
	if err != nil {
		log.Fatalln(err)
	}
	return
}

func (api *API) createTask(c *gin.Context) {
	res := []string{"Message", "Its works"}
	c.JSON(http.StatusOK, res)
	return
}

func (api *API) updateTask(c *gin.Context) {
	res := []string{"Message", "Its works"}
	c.JSON(http.StatusOK, res)
	return
}

func (api *API) getTask(c *gin.Context) {
	res := []string{"Message", "Its works"}
	c.JSON(http.StatusOK, res)
	return
}
func (api *API) deleteTask(c *gin.Context) {
	res := []string{"Message", "Its works"}
	c.JSON(http.StatusOK, res)
	return
}
