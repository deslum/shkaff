package fork

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/takama/daemon"
)

type serv struct {
	daemon.Daemon
}
type wrapped func()

func (service *serv) Run(function wrapped) (string, error) {
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
	function()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	killSignal := <-interrupt
	log.Println("Got signal:", killSignal)
	return "Service exited", nil
}

func InitDaemon() (service *serv, err error) {
	dependencies := []string{"shkaff.service"}
	srv, err := daemon.New("shkaff", "Backup database system", dependencies...)
	if err != nil {
		return
	}
	service = &serv{srv}
	return
}
