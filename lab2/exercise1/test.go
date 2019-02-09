package main

import (
	"ads/lab2"
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// TestNode is a wrapper for implementing lab2/exercise1
type TestNode struct {
	lab2.Node2
	responseCounter int
}

func (node *TestNode) onIncoming() {

	var incoming = make(chan lab2.Message)
	go node.Listen(incoming)

	// If it is a non-initiator, waits for the first message
	// in order to set the parent
	if !node.IsInitiator {
		message := <-incoming
		slice := strings.Split(message.Content, "#")

		node.Parent = slice[0]
		fmt.Println("Parent set to ", slice[0])

		node.Neighbours[slice[0]] = true
		node.responseCounter++
	}

	var sent = make(chan bool, len(node.Neighbours))
	for neighbour, visited := range node.Neighbours {
		if !visited {
			go node.SendMessage(node.ID, neighbour, sent)
			log.Debug("Message sent to ", neighbour)
		}
	}

	for {
		log.Debug("onIncoming - Status ", node)
		message := <-incoming

		slice := strings.Split(message.Content, "#")

		node.Neighbours[slice[0]] = true
		node.responseCounter++

		if len(node.Neighbours) == node.responseCounter {
			if node.IsInitiator {
				fmt.Println("It's decided! I'm ", node.ID)

			} else {
				fmt.Println("Responding to my parent! I'm ", node.ID)
				node.SendMessage(node.ID, node.Parent, sent)
			}
			break
		}
	}

	log.Debug("Waiting for messages to be sent...")
	for range node.Neighbours {
		<-sent
		log.Info("Add one sent!")
	}
}

func loadNode(filepath string) (node TestNode, err error) {
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
	node.ID = slice[0] + ":" + slice[1] + "#" + slice[2]
	node.IsInitiator = len(slice) > 3 && slice[3] == "*"
	node.responseCounter = 0

	node.Neighbours = make(map[string]bool)
	for scanner.Scan() {
		node.Neighbours[scanner.Text()] = false
	}
	err = scanner.Err()
	return node, err
}

func mainTest() {
	log.SetLevel(log.InfoLevel)
	log.Info("Running main test for node2...")

	if len(os.Args) < 2 {
		log.Error("Argument missing: configuration file. ")
		return
	}

	log.Info("Reading configuration file...")

	configDir := "ads/lab2/exercise1/files/echoConfig/"
	node, err := loadNode(configDir + os.Args[1])
	if err != nil {
		log.Error("Error creating the node!", err)
		return
	}

	node.onIncoming()
}
