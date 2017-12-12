package main

import (
	"log"
	"os"
	"shkaff/config"
	"shkaff/dashboard"
	"shkaff/fork"
	"shkaff/operator"
	"shkaff/statsender"
	"shkaff/worker"
)

type Creater interface {
	Init(action string, cfg config.ShkaffConfig) *Service
}
type Service interface {
	Run()
}

type shkaff struct{}

func (self *shkaff) Init(action string) (srv Service) {
	switch action {
	case "Operator":
		srv = operator.InitOperator()
	case "Worker":
		srv = worker.InitWorker()
	case "StatWorker":
		srv = statsender.InitStatSender()
	case "Dashboard":
		srv = dashboard.InitAPI()
	default:
		log.Fatalf("Unknown Shkaff service name %s\n", action)
	}
	return
}

func startShkaff() {
	servicesName := []string{"Operator", "Worker", "StatWorker", "Dashboard"}
	shkf := new(shkaff)
	for _, name := range servicesName {
		s := shkf.Init(name)
		go s.Run()
	}
}

func main() {
	daemon, err := fork.InitDaemon()
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
	status, err := daemon.Run(startShkaff)
	if err != nil {
		log.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	log.Println(status, err)
}
