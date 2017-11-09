package main

import (
	"log"
	"shkaff/config"
	"shkaff/operator"
)

type Creater interface {
	Init(action string, cfg config.ShkaffConfig) *Service
}
type Service interface {
	Run()
}

type Factory struct{}

func (self *Factory) Init(action string, cfg config.ShkaffConfig) Service {
	var srv Service
	switch action {
	case "Operator":
		srv = operator.InitOperator(cfg)
	default:
		log.Fatalf("Unknown Shkaff service name %s\n", action)
	}
	return srv
}

func main() {
	cfg := config.InitControlConfig()
	servicesName := []string{"Operator"}
	service := new(Factory)
	for _, name := range servicesName {
		srv := service.Init(name, cfg)
		srv.Run()
	}
}
