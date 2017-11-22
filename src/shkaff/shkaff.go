package main

import (
	"log"
	"os"
	"os/signal"
	"shkaff/config"
	"shkaff/operator"
	"shkaff/worker"
	"sync"
	"syscall"

	"github.com/takama/daemon"
)

type Creater interface {
	Init(action string, cfg config.ShkaffConfig) *Service
}
type Service interface {
	Run()
}

type serv struct {
	daemon.Daemon
}

type shkaff struct{}

func (self *shkaff) Init(action string) (srv Service) {
	switch action {
	case "Operator":
		srv = operator.InitOperator()
	case "Worker":
		srv = worker.InitWorker()
	default:
		log.Fatalf("Unknown Shkaff service name %s\n", action)
	}
	return
}

func (service *serv) start() (string, error) {
	usage := "Usage: shkaff install | remove | start | stop | status"
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	shkaffWG := sync.WaitGroup{}
	servicesName := []string{"Operator", "Worker"}
	shkf := new(shkaff)
	shkaffWG.Add(len(servicesName))
	for _, name := range servicesName {
		s := shkf.Init(name)
		go s.Run()
	}
	killSignal := <-interrupt
	log.Println("Got signal:", killSignal)
	return "Service exited", nil
}

func main() {
	dependencies := []string{"shkaff.service"}
	srv, err := daemon.New("shkaff", "Backup database system", dependencies...)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
	service := &serv{srv}
	status, err := service.start()
	if err != nil {
		log.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	log.Println(status, err)
}
