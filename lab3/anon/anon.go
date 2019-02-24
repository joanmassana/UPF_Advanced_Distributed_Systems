package main

import (
	"ads/lab3"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// TestNode is a wrapper for implementing 3/exercise2
type TestNode struct {
	lab3.Node
	responseCounter int
	largestId       int
	networkSize 	int
}

func (node *TestNode) stop(parent string, incoming chan lab3.Message) {
	message := buildMessage("stop", node, 0, 0)

	stopSent := make(chan bool, len(node.Neighbours))
	for neighbour := range node.Neighbours {
		if neighbour != parent {
			go node.SendMessage(message, neighbour, stopSent)
		}
	}

	for i := 0; i < len(node.Neighbours)-1; i++ {
		<-stopSent
		log.Debug("Stop sent!")
	}

	stoppedMessage := buildMessage("stopped", node, 0, 0)

	for i := 0; i < len(node.Neighbours)-1; {
		message := <-incoming
		if message.Content == "stopped" {
			i++
		} else if message.Content == "stop" {
			go node.SendMessage(stoppedMessage, message.OriginAddress+message.OriginPort, nil)
		}
	}

	if parent != "" {
		node.SendMessage(stoppedMessage, parent, nil)
	}
}

func (node *TestNode) onIncoming() {
	printNodeInfo(node)
	var incoming = make(chan lab3.Message)
	go node.Listen(incoming)

	var round = 0
	var subTreeSize = 0
	var nodesVisited = 0

	var lastMessage lab3.Message
	if node.IsInitiator {
		lastMessage = node.startWave(round)
	}

	for {
		incomingMessage := <-incoming

		if !node.IsInitiator && node.Parent == "" {
			subTreeSize = 0
			nodesVisited = 0
			lastMessage = node.joinWave(incomingMessage)

		} else {
			if incomingMessage.Round < lastMessage.Round {
				// do nothing
			} else if incomingMessage.Round > lastMessage.Round {
				subTreeSize = 0
				nodesVisited = 0
				lastMessage = node.joinWave(incomingMessage)
			} else {
				incomingID, _ := strconv.Atoi(incomingMessage.Content)
				if incomingID < lastMessage.ID {
					// do nothing
				} else if incomingID > lastMessage.ID {
					subTreeSize = 0
					nodesVisited = 0
					lastMessage = node.joinWave(incomingMessage)
				} else {
					subTreeSize += incomingMessage.SubTreeSize
					nodesVisited += 1
				}
			}
		}

		if !node.IsInitiator && nodesVisited == len(node.Neighbours) - 1 {
			node.sendToParent(incomingMessage, subTreeSize)

		} else if nodesVisited == len(node.Neighbours) {
			if subTreeSize == node.networkSize {
				// Leader
				fmt.Println("Leader")

			} else {
				subTreeSize = 0
				nodesVisited = 0
				round++
				node.startWave(round)
			}
		}
	}
}

func (node *TestNode) sendToParent(message lab3.Message, subTreeSize int) {
	messageToParent := buildMessage(message.Content, node, message.Round, subTreeSize + 1)

	var sent = make(chan bool)
	go node.SendMessage(messageToParent, node.Parent, sent)
	<-sent
}

func (node *TestNode) startWave(round int) (message lab3.Message) {
	node.setNeighborsToNotVisited()
	return node.sendToChildren(strconv.Itoa(node.ID), round, 0)
}

func (node *TestNode) joinWave(message lab3.Message) lab3.Message {
	node.setNeighborsToNotVisited()
	node.Parent = message.OriginAddress + message.OriginPort
	return node.sendToChildren(message.Content, message.Round, 0)
}

func (node *TestNode) sendToChildren(content string, round int, subTreeSize int) lab3.Message {
	message := buildMessage(content, node, round, subTreeSize)
	count := 0
	var sent = make(chan bool, len(node.Neighbours))
	for neighbour := range node.Neighbours {
		if neighbour != node.Parent {
			log.Debug("Sending from initiator to: " + neighbour)
			go node.SendMessage(message, neighbour, sent)
			count++
		}
	}
	for count > 0 {
		<-sent
		count--
	}
	return message
}

func (node *TestNode) setNeighborsToNotVisited() {
	for neighbour := range node.Neighbours {
		node.Neighbours[neighbour] = false
	}
}





func printNodeInfo(node *TestNode) {
	fmt.Println("Node Info ------------------------------")
	fmt.Printf("Self assigned ID -> %s          isInitiator -> %t\n", node.ID, node.IsInitiator)
	fmt.Printf("Address -> %s     Port -> %s\n", node.Address, node.Port)
	fmt.Println("Neighbors")
	for neighbour := range node.Neighbours {
		fmt.Printf(neighbour + ", ")
	}
	fmt.Printf("\n")
	fmt.Println("Parent ->", node.Parent)
	fmt.Println("----------------------------------------")
}

func buildMessage(content string, node *TestNode, round int, subTreeSize int) lab3.Message {
	message := lab3.Message{
		OriginAddress: 	node.Address,
		OriginPort:    	node.Port,
		ID:            	node.ID,
		Content:       	content,
		Error:        	nil,
		Round: 			round,
		SubTreeSize:	subTreeSize,

	}
	return message
}

func loadNode(filepath string, networkSize string) (node TestNode, err error) {
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
	node.IsInitiator = (len(slice) > 3 && slice[3] == "*" )|| (len(slice) > 2 && slice[2] == "*")
	node.responseCounter = 0
	node.largestId = 0
	node.networkSize, err = strconv.Atoi(networkSize)
	if err != nil {
		return node, err
	}
	node.ID = rand.Intn(node.networkSize)

	log.Debug("Setting neighbor map...")
	node.Neighbours = make(map[string]bool)
	for scanner.Scan() {
		node.Neighbours[scanner.Text()] = false

	}
	log.Debug("Neighbor map set...")
	err = scanner.Err()
	return node, err
}

func selfAssignRandomID() {

}

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("Running main test for node2...")

	if len(os.Args) < 3 {
		log.Error("Argument missing: configuration file. ")
		return
	}

	log.Info("Reading configuration file...")

	configDir := "ads/lab3/anon/files/configFiles/"
	node, err := loadNode(configDir + os.Args[1], os.Args[2])

	if err != nil {
		log.Error("Error creating the node!", err)
		return
	}

	node.onIncoming()
}
