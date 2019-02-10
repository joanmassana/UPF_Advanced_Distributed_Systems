package main

import (
	"bufio"
	"os"
	"strings"

	"ads/lab2"

	log "github.com/sirupsen/logrus"
)

func createNode(filepath string) (node lab2.Node, err error) {
	f, err := os.Open(filepath)
	if err != nil {
		return node, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	scanner.Scan()
	hostData := scanner.Text()
	slice := strings.Split(hostData, ":")
	node.Port = ":" + slice[1]
	node.ID = slice[2]
	if len(slice) > 3 {
		node.IsInitiator = slice[3] == "*"
	}


	node.Neighbours = make(map[string]bool)
	for scanner.Scan() {
		node.Neighbours[scanner.Text()] = false
	}
	err = scanner.Err()
	return node, err
}

func main() {

	log.SetLevel(log.DebugLevel)

	if len(os.Args) < 2 {
		log.Error("Argument missing: configuration file. ")
		return
	}

	log.Info("Reading configuration file...")

	node, err := createNode(os.Args[1])
	if err != nil {
		log.Error("Error creating the node!", err)
		return
	}

	node.StopChannel = make(chan bool)

	go node.Start()

	<-node.StopChannel

}
