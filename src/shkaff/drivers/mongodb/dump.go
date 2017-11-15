package mongodb

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"shkaff/structs"
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

func InitDriver(task *structs.Task) (mp *MongoParams) {
	return &MongoParams{
		host:                   task.Host,
		port:                   task.Port,
		login:                  task.DBUser,
		password:               task.DBPassword,
		ipv6:                   task.Ipv6,
		gzip:                   task.Gzip,
		database:               task.Database,
		collection:             task.Sheet,
		parallelCollectionsNum: task.ThreadCount,
	}
}

func (mp *MongoParams) Dump() {
	var stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", mp.ParamsToString())
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println(stderr.String())
}
