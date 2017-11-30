package statsender

import (
	"encoding/json"
	"log"
	"shkaff/drivers/rmq/producer"
	"shkaff/structs"
)

type StatSender struct {
	sChan    chan structs.StatMessage
	producer *producer.RMQ
}

func Run() (stat *StatSender) {
	stat = new(StatSender)
	stat.sChan = make(chan structs.StatMessage)
	stat.producer = producer.InitAMQPProducer("shkaff_stat")
	go stat.statSender()
	return
}

func (stat *StatSender) SendStatMessage(action structs.Action, userID, dbid, taskID int, err error) {
	var statMessage structs.StatMessage
	statMessage.Act = action
	statMessage.UserID = userID
	statMessage.DBID = dbid
	statMessage.TaskID = taskID
	statMessage.Error = err
	stat.sChan <- statMessage
}

func (stat *StatSender) statSender() {
	for {
		select {
		case statMsg := <-stat.sChan:
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
}
