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
	networkSize     int
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
	var lastMessageID int
	if node.IsInitiator {
		lastMessage = node.startWave(round)
		lastMessageID = node.ID
	}

	loopCount := 0
	for loopCount < 10 {
		loopCount++

		incomingMessage := <-incoming

		fmt.Println("%v", incomingMessage)
		if !node.IsInitiator && node.Parent == "" {
			log.Info("JoinWave 1")

			subTreeSize = 0
			nodesVisited = 0
			lastMessage = node.joinWave(incomingMessage)
			lastMessageID, _ = strconv.Atoi(lastMessage.Content)

		} else {
			if incomingMessage.Round < lastMessage.Round {
				// do nothing
			} else if incomingMessage.Round > lastMessage.Round {
				log.Info("JoinWave 2")
				subTreeSize = 0
				nodesVisited = 0
				lastMessage = node.joinWave(incomingMessage)
				lastMessageID, _ = strconv.Atoi(lastMessage.Content)
			} else {
				incomingID, _ := strconv.Atoi(incomingMessage.Content)
				if incomingID < lastMessageID {
					// do nothing
				} else if incomingID > lastMessageID {
					log.Info("JoinWave 3")
					subTreeSize = 0
					nodesVisited = 0
					lastMessage = node.joinWave(incomingMessage)
					lastMessageID, _ = strconv.Atoi(lastMessage.Content)
				} else {
					subTreeSize += incomingMessage.SubTreeSize
					nodesVisited++
				}
			}
		}

		fmt.Println("Nodes visited: ", nodesVisited)
		fmt.Println("Size: ", subTreeSize)
		if lastMessageID != node.ID && nodesVisited == len(node.Neighbours)-1 {
			node.sendToParent(incomingMessage, subTreeSize)

		} else if nodesVisited == len(node.Neighbours) {
			if subTreeSize == node.networkSize-1 {
				// Leader
				fmt.Println("Leader!")

			} else {
				subTreeSize = 0
				nodesVisited = 0
				round++
				node.ID = rand.Intn(node.networkSize)
				lastMessage = node.startWave(round)
				lastMessageID, _ = strconv.Atoi(lastMessage.Content)
			}
		}
	}
}

func (node *TestNode) sendToParent(message lab3.Message, subTreeSize int) {
	messageToParent := buildMessage(message.Content, node, message.Round, subTreeSize+1)

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
	log.Debug("Parent set to ", node.Parent)

	return node.sendToChildren(message.Content, message.Round, 0)
}

func (node *TestNode) sendToChildren(content string, round int, subTreeSize int) lab3.Message {
	message := buildMessage(content, node, round, subTreeSize)
	count := 0
	var sent = make(chan bool, len(node.Neighbours))
	for neighbour := range node.Neighbours {
		if neighbour != node.Parent {
			log.Debug("Sending from ", node.Port, " to: ", neighbour)
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
	fmt.Printf("Self assigned ID -> %v          isInitiator -> %t\n", node.ID, node.IsInitiator)
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
		OriginAddress: node.Address,
		OriginPort:    node.Port,
		ID:            node.ID,
		Content:       content,
		Error:         nil,
		Round:         round,
		SubTreeSize:   subTreeSize,
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
	node.IsInitiator = (len(slice) > 3 && slice[3] == "*") || (len(slice) > 2 && slice[2] == "*")
	node.responseCounter = 0
	node.networkSize, err = strconv.Atoi(networkSize)
	if err != nil {
		return node, err
	}
	seed, _ := strconv.ParseInt(slice[1], 0, 0)
	rand.Seed(seed)
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
	log.SetLevel(log.InfoLevel)
	log.Info("Running main test for node2...")

	if len(os.Args) < 3 {
		log.Error("Argument missing: configuration file. ")
		return
	}

	log.Info("Reading configuration file...")

	configDir := "ads/lab3/anon/files/configFiles/"
	node, err := loadNode(configDir+os.Args[1], os.Args[2])

	if err != nil {
		log.Error("Error creating the node!", err)
		return
	}

	node.onIncoming()
}
