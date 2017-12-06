package statsender

import (
	"encoding/json"
	"log"
	"shkaff/drivers/rmq/consumer"
	"shkaff/drivers/rmq/producer"
	"shkaff/drivers/stat"
	"shkaff/structs"
	"time"
)

type StatSender struct {
	sChan    chan structs.StatMessage
	producer *producer.RMQ
	consumer *consumer.RMQ
	statDB   *stat.StatDB
}

func Run() (statSender *StatSender) {
	log.Println("Start StatSender")
	statSender = new(StatSender)
	statSender.sChan = make(chan structs.StatMessage)
	statSender.producer = producer.InitAMQPProducer("shkaff_stat")
	statSender.consumer = consumer.InitAMQPConsumer()
	statSender.statDB = stat.InitStat()
	go statSender.statSender()
	go statSender.statWorker()
	return
}

func (statSender *StatSender) SendStatMessage(action structs.Action, userID, dbid, taskID int, err error) {
	var statMessage structs.StatMessage
	statMessage.UserID = uint16(userID)
	statMessage.DbID = uint16(dbid)
	statMessage.TaskID = uint16(taskID)
	statMessage.CreateDate = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	switch action {
	case 0:
		statMessage.NewOperator = 1
	case 1:
		statMessage.SuccessOperator = 1
	case 2:
		statMessage.FailOperator = 1
		statMessage.ErrorOperator = err.Error()
	case 3:
		statMessage.NewDump = 1
	case 4:
		statMessage.SuccessDump = 1
	case 5:
		statMessage.FailDump = 1
		statMessage.ErrorDump = err.Error()
	case 6:
		statMessage.NewRestore = 1
	case 7:
		statMessage.SuccessRestore = 1
	case 8:
		statMessage.FailRestore = 1
		statMessage.ErrorRestore = err.Error()
	}
	statSender.sChan <- statMessage
}

func (statSender *StatSender) statSender() {
	for {
		select {
		case statMsg := <-statSender.sChan:
			msg, err := json.Marshal(statMsg)
			if err != nil {
				log.Println(err)
				continue
			}
			err = statSender.producer.Publish(msg)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func (statSender *StatSender) statWorker() {
	var statMessage structs.StatMessage
	statSender.consumer.InitConnection("shkaff_stat")
	for message := range statSender.consumer.Msgs {
		err := json.Unmarshal(message.Body, &statMessage)
		if err != nil {
			log.Println("statWorker", err, "Failed JSON parse")
			continue
		}
		err = statSender.statDB.Insert(statMessage)
		if err != nil {
			log.Println("statWorker", err)
			continue
		}
		message.Ack(false)
	}
}
