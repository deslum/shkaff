package mongodb

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"shkaff/structs"
	"shkaff/structs/databases"
	"strings"
)

const (
	dumpCommand            = "mongodump"
	hostKey                = "--host"
	portKey                = "--port"
	loginKey               = "--username"
	passKey                = "--password"
	ipv6Key                = "--ipv6"
	databaseKey            = "--db"
	collectionKey          = "--collection"
	gzipKey                = "--gzip"
	parallelCollectionsKey = "-j"
)

type MongoParams struct {
	host                   string
	port                   int
	login                  string
	password               string
	ipv6                   bool
	database               string
	collection             string
	gzip                   bool
	parallelCollectionsNum int
}

func (mp *MongoParams) isUseAuth() bool {
	return mp.login != "" && mp.password != ""
}

func (mp *MongoParams) ParamsToString() (commandString string) {
	var cmdLine []string
	cmdLine = append(cmdLine, dumpCommand)
	cmdLine = append(cmdLine, fmt.Sprintf("%s %s", hostKey, mp.host))
	cmdLine = append(cmdLine, fmt.Sprintf("%s %d", portKey, mp.port))
	if mp.isUseAuth() {
		auth := fmt.Sprintf("%s %s %s %s", loginKey, mp.login, passKey, mp.password)
		cmdLine = append(cmdLine, auth)
	}
	if mp.ipv6 {
		cmdLine = append(cmdLine, ipv6Key)
	}
	if mp.gzip {
		cmdLine = append(cmdLine, gzipKey)
	}
	if mp.database != "" {
		cmdLine = append(cmdLine, fmt.Sprintf("%s=%s", databaseKey, mp.database))
		if mp.collection != "" {
			cmdLine = append(cmdLine, fmt.Sprintf("%s=%s", collectionKey, mp.collection))
		}
	}
	if mp.collection == "" && mp.parallelCollectionsNum > 4 {
		cmdLine = append(cmdLine, fmt.Sprintf("%s=%d", parallelCollectionsKey, mp.parallelCollectionsNum))
	}
	commandString = strings.Join(cmdLine, " ")
	return
}

func InitDriver() (mp databases.DatabaseDriver) {
	return &MongoParams{}
}

func (mp *MongoParams) setDBSettings(task *structs.Task) {
	mp.host = task.Host
	mp.port = task.Port
	mp.login = task.DBUser
	mp.password = task.DBPassword
	mp.ipv6 = task.Ipv6
	mp.gzip = task.Gzip
	mp.database = task.Database
	mp.collection = task.Sheet
	mp.parallelCollectionsNum = task.ThreadCount
}

func (mp *MongoParams) Dump(task *structs.Task) {
	var stderr bytes.Buffer
	mp.setDBSettings(task)
	cmd := exec.Command("sh", "-c", mp.ParamsToString())
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	log.Println(stderr.String())
}
