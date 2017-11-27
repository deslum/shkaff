package worker

import (
	"encoding/json"
	"log"
	"shkaff/drivers/rmq/producer"
	"shkaff/structs"
)

type statSender struct {
	sChan    chan structs.StatMessage
	producer *producer.RMQ
}

func startStatSender(sChan chan structs.StatMessage) {
	stat := new(statSender)
	stat.sChan = make(chan structs.StatMessage, 100)
	stat.producer = producer.InitAMQPProducer("shkaff_stat")
	
	go stat.statSender()

}
func (stat *statSender) statSender() {
	for {
		statMsg, ok := <-stat.sChan
		if !ok {
			break
		}
		msg, err := json.Marshal(statMsg)
		if err != nil {
			log.Println(err)
			continue
		}
		err = stat.producer.Publish(msg)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
