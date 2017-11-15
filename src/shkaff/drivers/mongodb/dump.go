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
	DUMP_COMMAND            = "mongodump"
	HOST_KEY                = "--host"
	PORT_KEY                = "--port"
	LOGIN_KEY               = "--username"
	PASS_KEY                = "--password"
	IPV6_KEY                = "--ipv6"
	DATABASE_KEY            = "--db"
	COLLECTION_KEY          = "--collection"
	GZIP_KEY                = "--gzip"
	PARALLEL_KEY 			= "-j"
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
	cmdLine = append(cmdLine, DUMP_COMMAND)
	cmdLine = append(cmdLine, fmt.Sprintf("%s %s", HOST_KEY, mp.host))
	cmdLine = append(cmdLine, fmt.Sprintf("%s %d", PORT_KEY, mp.port))
	if mp.isUseAuth() {
		auth := fmt.Sprintf("%s %s %s %s", LOGIN_KEY, mp.login, PASS_KEY, mp.password)
		cmdLine = append(cmdLine, auth)
	}
	if mp.ipv6 {
		cmdLine = append(cmdLine, IPV6_KEY)
	}
	if mp.gzip {
		cmdLine = append(cmdLine, GZIP_KEY)
	}
	if mp.database != "" {
		cmdLine = append(cmdLine, fmt.Sprintf("%s=%s", DATABASE_KEY, mp.database))
		if mp.collection != "" {
			cmdLine = append(cmdLine, fmt.Sprintf("%s=%s", COLLECTION_KEY, mp.collection))
		}
	}
	if mp.collection == "" && mp.parallelCollectionsNum > 4 {
		cmdLine = append(cmdLine, fmt.Sprintf("%s=%d", PARALLEL_KEY, mp.parallelCollectionsNum))
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
