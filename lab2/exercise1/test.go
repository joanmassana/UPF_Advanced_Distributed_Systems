package main

import (
	"ads/lab2"
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// TestNode is a wrapper for implementing lab2/exercise1
type TestNode struct {
	lab2.Node2
	responseCounter int
}

func (node *TestNode) onIncoming() {
	printNodeInfo(node)
	var incoming = make(chan lab2.Message)
	go node.Listen(incoming)

	// If it is a non-initiator, waits for the first message
	// in order to set the parent
	if !node.IsInitiator {

		message := <-incoming

		node.Parent = message.OriginAddress + message.OriginPort
		fmt.Println("My parent is", message.ID + " - " + message.OriginAddress + message.OriginPort)

		node.Neighbours[message.OriginAddress+message.OriginPort] = true
		node.responseCounter++
	}

	var sent = make(chan bool, len(node.Neighbours))
	for neighbour, visited := range node.Neighbours {
		if !visited {
			message := buildMessage("", node)
			log.Debug("Sending to: " + neighbour)
			go node.SendMessage(message, neighbour, sent)
		}
	}

	for {
		log.Debug("onIncoming - Status ", node)
		log.Debug("neighbors total", len(node.Neighbours))
		message := <-incoming

		node.Neighbours[message.OriginAddress+message.OriginPort] = true
		node.responseCounter++

		if len(node.Neighbours) == node.responseCounter {
			if node.IsInitiator {
				fmt.Println("Decision Event: I'm", node.ID)

			} else {
				fmt.Println("To parent: I'm", node.ID)
				message := buildMessage("", node)
				node.SendMessage(message, node.Parent, sent)
			}
			break
		}
	}

	log.Debug("Waiting for messages to be sent...")
	for range node.Neighbours {
		<-sent
		log.Debug("Add one sent!")
	}
}

func printNodeInfo(node *TestNode) {
	fmt.Println("Node Info ------------------------------")
	fmt.Printf("ID -> %s          isInitiator -> %t\n", node.ID, node.IsInitiator)
	fmt.Printf("Address -> %s     Port -> %s\n", node.Address, node.Port)
	fmt.Println("Neighbors")
	for neighbour := range node.Neighbours {
		fmt.Printf(neighbour + ", ")
	}
	fmt.Printf("\n")
	fmt.Println("Parent ->", node.Parent)
	fmt.Println("----------------------------------------")
}

func buildMessage(content string, node *TestNode) lab2.Message {
	message := lab2.Message{
		OriginAddress: node.Address,
		OriginPort:    node.Port,
		ID:            node.ID,
		Content:       content,
		Error:         nil}
	return message
}

func loadNode(filepath string) (node TestNode, err error) {
	log.Debug("Loading node...")
	f, err := os.Open(filepath)
	if err != nil {
		return node, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	scanner.Scan()

	hostData := scanner.Text()
	slice := strings.Split(hostData, ":")

	log.Debug("Init data in node...")

	node.Address = slice[0]
	node.Port = ":" + slice[1]
	node.ID = slice[2]
	node.IsInitiator = len(slice) > 3 && slice[3] == "*"
	node.responseCounter = 0

	log.Debug("Setting neighbor map...")
	node.Neighbours = make(map[string]bool)
	for scanner.Scan() {
		node.Neighbours[scanner.Text()] = false

	}
	log.Debug("Neighbor map set...")
	err = scanner.Err()
	return node, err
}

func main() {
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
