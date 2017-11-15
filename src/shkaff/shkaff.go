package main

import (
	"log"
	"shkaff/config"
	"shkaff/operator"
	"shkaff/worker"
	"github.com/sevlyar/go-daemon"
	"sync"
	"os"
	"syscall"
	"flag"
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

var (
	signal = flag.String("s", "", `send signal to the daemon
		quit — graceful shutdown
		stop — fast shutdown`)
	stop = make(chan struct{})
	done = make(chan struct{})
)

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}

func start(){
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

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)
	if err := os.MkdirAll("./logs", 0644); err!=nil{
		log.Fatalln(err)
	}
	cntxt := &daemon.Context{
		PidFileName: "/tmp/shkaff.pid",
		PidFilePerm: 0644,
		LogFileName: "./logs/shkaff.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalln("Unable send signal to the daemon:", err)
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()
	log.Println("Shkaff daemon started")
	go start()
	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("Shkaff daemon terminated")
	
}
