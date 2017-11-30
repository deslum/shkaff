package mongodb

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"shkaff/consts"
	"shkaff/structs"
	"shkaff/structs/databases"
	"strings"
)

var (
	MONGO_SUCESS_DUMP    = regexp.MustCompile(`\tdone\ dumping`)
	MONGO_SUCESS_RESTORE = regexp.MustCompile(`\tdone\ dumping`)
)

func (mp *MongoParams) isUseAuth() bool {
	return mp.login != "" && mp.password != ""
}

func (mp *MongoParams) ParamsToString() (commandString string) {
	var cmdLine []string
	cmdLine = append(cmdLine, consts.MONGO_DUMP_COMMAND)
	cmdLine = append(cmdLine, fmt.Sprintf("%s %s", consts.MONGO_HOST_KEY, mp.host))
	cmdLine = append(cmdLine, fmt.Sprintf("%s %d", consts.MONGO_PORT_KEY, mp.port))
	if mp.isUseAuth() {
		auth := fmt.Sprintf("%s %s %s %s", consts.MONGO_LOGIN_KEY, mp.login, consts.MONGO_PASS_KEY, mp.password)
		cmdLine = append(cmdLine, auth)
	}
	if mp.ipv6 {
		cmdLine = append(cmdLine, consts.MONGO_GZIP_KEY)
	}
	if mp.gzip {
		cmdLine = append(cmdLine, consts.MONGO_GZIP_KEY)
	}
	if mp.database != "" {
		cmdLine = append(cmdLine, fmt.Sprintf("%s=%s", consts.MONGO_DATABASE_KEY, mp.database))
		if mp.collection != "" {
			cmdLine = append(cmdLine, fmt.Sprintf("%s=%s", consts.MONGO_COLLECTION_KEY, mp.collection))
		}
	}
	if mp.collection == "" && mp.parallelCollectionsNum > 4 {
		cmdLine = append(cmdLine, fmt.Sprintf("%s=%d", consts.MONGO_PARALLEL_KEY, mp.parallelCollectionsNum))
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

func (mp *MongoParams) Dump(task *structs.Task) (dumpResult string, err error) {
	var stderr bytes.Buffer
	mp.setDBSettings(task)
	cmd := exec.Command("sh", "-c", mp.ParamsToString())
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	}
	dumpResult = stderr.String()
	reResult := MONGO_SUCESS_DUMP.FindString(dumpResult)
	if reResult != "" {
		return
	}
	return dumpResult, errors.New("Bad dump result")
}

func (mp *MongoParams) Restore(task *structs.Task) (restoreResult string, err error) {
	log.Println("Mock restore")
	// var stderr bytes.Buffer
	// mp.setDBSettings(task)
	// cmd := exec.Command("sh", "-c", mp.ParamsToString())
	// cmd.Stderr = &stderr
	// if err := cmd.Run(); err != nil {
	// 	log.Println(fmt.Sprint(err) + ": " + stderr.String())
	// 	return "", err
	// }
	// restoreResult = stderr.String()
	// reResult := MONGO_SUCESS_DUMP.FindString(restoreResult)
	// if reResult != "" {
	// 	return
	// }
	return restoreResult, errors.New("Bad restore result")
}
