package main

import (
	"bufio"
	"os"
	"strings"

	lab1 "ads/lab1"

	log "github.com/sirupsen/logrus"
)

func createNode(filepath string) (node lab1.Node, err error) {
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

	for scanner.Scan() {
		node.Neighbours = append(node.Neighbours, scanner.Text())
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
