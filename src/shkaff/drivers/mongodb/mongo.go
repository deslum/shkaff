package mongodb

import (
	"encoding/json"
	"fmt"
	"log"
	"shkaff/structs"
	"time"

	"gopkg.in/mgo.v2"
)

type mongoCliStruct struct {
	task     structs.Task
	messages []structs.Task
}

func (m *mongoCliStruct) forEmptyDatabases() {
	url := fmt.Sprintf("%s:%d", m.task.Host, m.task.Port)
	session, err := mgo.DialWithTimeout(url, 5*time.Second)
	if err != nil {
		log.Println(err)
		return
	}
	defer session.Close()

	dbNames, err := session.DatabaseNames()
	if err != nil {
		log.Println(err)
		return
	}
	for _, dbName := range dbNames {
		m.task.Database = dbName
		m.messages = append(m.messages, m.task)
	}
	return
}

func (m *mongoCliStruct) forFillDatabases() {
	databases := make(map[string][]string)
	err := json.Unmarshal([]byte(m.task.Databases), &databases)
	if err != nil {
		log.Println("Error unmarshal databases", databases, err)
		return
	}
	for base := range databases {
		m.task.Database = base
		m.messages = append(m.messages, m.task)
	}
	return
}

func GetMessages(task structs.Task) (caches []structs.Task) {
	var mongo = new(mongoCliStruct)
	mongo.task = task
	if task.Databases == "{}" {
		mongo.forEmptyDatabases()
	} else {
		mongo.forFillDatabases()
	}
	return mongo.messages

}
