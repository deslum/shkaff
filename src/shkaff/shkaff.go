package main

import (
	"log"
	"shkaff/config"
	"shkaff/operator"
	"shkaff/worker"
	"sync"
)

type Creater interface {
	Init(action string, cfg config.ShkaffConfig) *Service
}
type Service interface {
	Run()
}

type shkaff struct{}

func (self *shkaff) Init(action string, cfg config.ShkaffConfig) (srv Service) {
	switch action {
	case "Operator":
		srv = operator.InitOperator(cfg)
	case "Worker":
		srv = worker.InitWorker(cfg)
	default:
		log.Fatalf("Unknown Shkaff service name %s\n", action)
	}
	return
}

func main() {
	var serviceCount int
	shkaffWG := sync.WaitGroup{}
	cfg := config.InitControlConfig()
	servicesName := []string{"Operator", "Worker"}
	service := new(shkaff)
	for _, name := range servicesName {
		var s Service
		serviceCount++
		shkaffWG.Add(serviceCount)
		s = service.Init(name, cfg)
		go s.Run()
	}
	shkaffWG.Wait()
}
