package worker

import "fmt"

func dumpAnalyser(dchan chan string) {
	for {
		msgStr, ok := <-dchan
		if !ok {
			break
		}
		fmt.Println(msgStr)
	}
}
