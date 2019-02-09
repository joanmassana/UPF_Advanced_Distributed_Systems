package main

import (
	lab1 "ads/lab1"
	"bufio"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// TestNode is a wrapper for testing Node2
type TestNode struct {
	lab1.Node2
}

func (node *TestNode) onIncoming(stopChannel chan bool) {

	var incoming = make(chan lab1.Message)
	go node.Listen(incoming)

	for {
		message := <-incoming
		// node.SendToAllNeighbours(message.Content, nil)

		if message.Error != nil {
			log.Error("Error while receiving incoming message: ", message.Error)
		} else {
			fmt.Println(message.Content)
			if message.Content == "stop" {
				break
			}
		}
	}
	stopChannel <- true
}

func (node *TestNode) onUserInput(stopChannel chan bool) {

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Input full filename of file to send: ")
		data, readError := reader.ReadString('\n')

		if readError != nil {
			log.Error("onUserInput - Error reading std input: ", readError)
			continue
		}

		text := lab1.EraseNewLines(data)
		log.Info("onUserInput - Text to send: '", text, "'")

		if text == "" {
			// do nothing
			continue
		}

		stopSent := make(chan bool)
		node.SendToAllNeighbours(text, stopSent)

		if text == "stop" {
			for range node.Neighbours {
				log.Debug("onUserInput - Waiting stop sent...")
				<-stopSent
				log.Debug("onUserInput - Stop sent!")
			}
			break
		}
	}
	stopChannel <- true
}

func mainTest() {

	log.SetLevel(log.DebugLevel)
	log.Info("Running main test for node2...")

	otherIP := "127.0.0.1"
	otherPort := ":6002"
	thisPort := ":6001"
	if len(os.Args) > 1 {
		fmt.Println("Ports set by arguments...")
		thisPort = os.Args[1]
		otherPort = os.Args[2]
	}

	node := TestNode{
		Node2: lab1.Node2{
			Port: thisPort,
			Neighbours: []string{
				otherIP + otherPort,
			},
		},
	}

	stopChannel := make(chan bool)
	go node.onIncoming(stopChannel)
	go node.onUserInput(stopChannel)

	<-stopChannel
}
