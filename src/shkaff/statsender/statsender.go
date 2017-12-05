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

type Stat struct {
	UserID          uint16    `db:"UserId" json:"uid"`
	DbID            uint16    `db:"DbID" json:"did"`
	TaskID          uint16    `db:"TaskId" json:"tid"`
	NewOperator     uint32    `db:"NewOperator" json:"no"`
	SuccessOperator uint32    `db:"SuccessOperator" json:"so"`
	FailOperator    uint32    `db:"FailOperator" json:"fo"`
	ErrorOperator   string    `db:"ErrorOperator" json:"eo"`
	NewDump         uint32    `db:"NewDump" json:"nd"`
	SuccessDump     uint32    `db:"SuccessDump" json:"sd"`
	FailDump        uint32    `db:"FailDump" json:"fd"`
	ErrorDump       string    `db:"ErrorDump" json:"ed"`
	NewRestore      uint32    `db:"NewRestore" json:"nr"`
	SuccessRestore  uint32    `db:"SuccessRestore" json:"sr"`
	FailRestore     uint32    `db:"FailRestore" json:"fr"`
	ErrorRestore    string    `db:"ErrorRestore" json:"er"`
	CreateDate      time.Time `db:"CreateDate" json:"cd"`
}

type StatSender struct {
	sChan    chan Stat
	producer *producer.RMQ
	consumer *consumer.RMQ
	statDB   *stat.StatDB
}

func Run() (statSender *StatSender) {
	log.Println("Start StatSender")
	statSender = new(StatSender)
	statSender.sChan = make(chan Stat)
	statSender.producer = producer.InitAMQPProducer("shkaff_stat")
	statSender.consumer = consumer.InitAMQPConsumer()
	statSender.statDB = stat.InitStat()
	go statSender.statSender()
	go statSender.statWorker()
	return
}

func (statSender *StatSender) SendStatMessage(action structs.Action, userID, dbid, taskID int, err error) {
	var statMessage Stat
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
	var statMessage Stat
	db := statSender.statDB.DB
	statSender.consumer.InitConnection("shkaff_stat")
	for message := range statSender.consumer.Msgs {
		err := json.Unmarshal(message.Body, &statMessage)
		if err != nil {
			log.Println("statWorker", err, "Failed JSON parse")
			continue
		}
		tx := db.MustBegin()
		_, err = tx.NamedExec("INSERT INTO shkaff_stat (UserId, DbID, TaskId, NewOperator, SuccessOperator, FailOperator, ErrorOperator, NewDump, SuccessDump, FailDump, ErrorDump, NewRestore, SuccessRestore, FailRestore, ErrorRestore, CreateDate) VALUES (:UserId, :DbID, :TaskId, :NewOperator, :SuccessOperator, :FailOperator, :ErrorOperator, :NewDump, :SuccessDump, :FailDump, :ErrorDump, :NewRestore, :SuccessRestore, :FailRestore, :ErrorRestore, :CreateDate)", &statMessage)
		if err != nil {
			log.Println(err)
			continue
		}
		if err := tx.Commit(); err != nil {
			log.Println(err)
			continue
		}
		message.Ack(false)
	}
}
