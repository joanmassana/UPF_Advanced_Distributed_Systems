package main

import (
	"os"
	"strings"

	lab1 "ads/lab1"

	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetLevel(log.DebugLevel)

	otherIP := "127.0.0.1"
	otherPort := ":6002"
	thisPort := ":6001"

	if len(os.Args) > 1 {
		log.Info("Ports set by arguments...")

		thisPort = os.Args[1]
		if !strings.HasPrefix(":", thisPort) {
			thisPort = ":" + thisPort
		}

		otherPort = os.Args[2]
		if !strings.HasPrefix(":", otherPort) {
			otherPort = ":" + otherPort
		}
	}

	stopChannel := make(chan bool)

	node := lab1.Node{
		Port: thisPort,
		Neighbours: []string{
			otherIP + otherPort,
		},
		StopChannel: stopChannel,
	}
	go node.Start()

	<-stopChannel

}
