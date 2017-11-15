package databases

import (
	"log"
	"shkaff/drivers/mongodb"
	"shkaff/structs"
)

type Creater interface {
	Init() *Service
}

type Service interface {
	New()
	Dump()
}

type Driver struct{}

func Init(database string, task *structs.Task) (driver Driver) {
	switch database {
	case "mongodb":
		driver = mongodb.InitDriver(task)
	default:
		log.Fatalf("Unknown driver for %s\n", database)
	}
	return
}

// func main() {
// 	var serviceCount int
// 	shkaffWG := sync.WaitGroup{}
// 	cfg := config.InitControlConfig()
// 	servicesName := []string{"Operator", "Worker"}
// 	service := new(shkaff)
// 	for _, name := range servicesName {
// 		var s Service
// 		serviceCount++
// 		shkaffWG.Add(serviceCount)
// 		s = service.Init(name, cfg)
// 		go s.Run()
// 	}
// 	shkaffWG.Wait()
// }
