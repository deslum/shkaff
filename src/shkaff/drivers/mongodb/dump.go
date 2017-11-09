package mongodb

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

const (
	dumpCommand            = "mongodump"
	hostKey                = "--host"
	portKey                = "--port"
	loginKey               = "--login"
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

func New(host string, port int, login, password string, ipv6 bool, database, collection string,
	gzip bool, parallelCollectionsNum int) (mp MongoParams) {

	if err := net.ParseIP(host); err != nil {
		log.Fatalf("MongoDB Host %s invalid \n", host)
	}
	if port < 1024 || port > 65535 {
		log.Fatalf("MongoDB Port %d invalid \n", port)
	}
	if (login == "" && password != "") || (login != "" && password == "") {
		log.Fatalf("MongoDB bad authorization \n")
	}
	return MongoParams{
		host:                   host,
		port:                   port,
		login:                  login,
		password:               password,
		ipv6:                   ipv6,
		gzip:                   gzip,
		database:               database,
		collection:             collection,
		parallelCollectionsNum: parallelCollectionsNum,
	}
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
