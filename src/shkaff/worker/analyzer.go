package worker

import (
	"log"
	"regexp"
)

var (
	MONGO_SUCESS_DUMP = regexp.MustCompile(`\tdone\ dumping`)
)

func dumpAnalyser(dchan chan string) {
	for {
		msgStr, ok := <-dchan
		if !ok {
			break
		}
		reResult := MONGO_SUCESS_DUMP.FindString(msgStr)
		if reResult != "" {
			log.Println("OK")
		}
	}
}
