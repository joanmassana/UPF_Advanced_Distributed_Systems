package main

import (
	"ads/lab2"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// TestNode is a wrapper for implementing lab2/exercise2
type TestNode struct {
	lab2.Node
	responseCounter int
	largestId       string
}

func (node *TestNode) onIncoming() {
	printNodeInfo(node)
	var incoming = make(chan lab2.Message)
	go node.Listen(incoming)
	count := 0

	if node.IsInitiator {
		node.largestId = node.ID
		var sentFromInit = make(chan bool, len(node.Neighbours))
		for neighbour := range node.Neighbours {
			message := buildMessage(node.largestId, node, count)
			count++
			log.Debug("Sending from init to: " + neighbour)
			go node.SendMessage(message, neighbour, sentFromInit)
		}
		for range node.Neighbours {
			<-sentFromInit
		}
	}

	var sent = make(chan bool, len(node.Neighbours))
	for {
		//Wait for message
		message := <-incoming

		log.Debug("Incoming Message. Content is" + message.Content + ". Our largestId is" + node.largestId)
		if message.Content == "stop" {
			return
		}
		content, _ := strconv.Atoi(message.Content)
		largestId, _ := strconv.Atoi(node.largestId)

		if content > largestId {
			log.Debug("Message ID larger")
			node.largestId = message.Content
			log.Debug("LargestId set to " + node.largestId)
			//reset neighbors to not sent
			for neighbour := range node.Neighbours {
				node.Neighbours[neighbour] = false
			}
			//reset sent channels
			sent = make(chan bool, len(node.Neighbours))
			node.responseCounter = 0
			//Set sender as parent
			node.Parent = message.OriginAddress + message.OriginPort
			log.Debug(node.Parent + " is now my parent")
			node.Neighbours[message.OriginAddress+message.OriginPort] = true
			node.responseCounter++
			//send messages to neighbors with new largest id
			for neighbour, visited := range node.Neighbours {
				if !visited {
					message := buildMessage(node.largestId, node, count)
					count++
					go node.SendMessage(message, neighbour, sent)
				}
			}
			if len(node.Neighbours) == node.responseCounter {
				message := buildMessage(node.largestId, node, count)
				count++
				node.SendMessage(message, node.Parent, sent)

			}
		} else if content < largestId {
			//Do nothing
			log.Debug("Message ID smaller")

		} else {
			log.Debug("Message ID equal")
			node.Neighbours[message.OriginAddress+message.OriginPort] = true
			node.responseCounter++

			if len(node.Neighbours) == node.responseCounter {
				if node.IsInitiator && node.ID == node.largestId {
					fmt.Println("---- DECISION EVENT ----> PROCESS w/ ID #" + node.ID + ": I'm leader!")
					message := buildMessage("stop", node, count)

					stopSent := make(chan bool, len(node.Neighbours))
					node.SendToAllNeighbours(message, stopSent)
					for range node.Neighbours {
						<-stopSent
						log.Debug("Add one sent!")
					}

					return
				} else {
					message := buildMessage(node.largestId, node, count)
					count++
					node.SendMessage(message, node.Parent, sent)
				}

			}
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

func buildMessage(content string, node *TestNode, count int) lab2.Message {
	message := lab2.Message{
		OriginAddress: node.Address,
		OriginPort:    node.Port,
		ID:            node.ID,
		Num:           count,
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
	node.largestId = "0"

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
	log.SetLevel(log.DebugLevel)
	log.Info("Running main test for node2...")

	if len(os.Args) < 2 {
		log.Error("Argument missing: configuration file. ")
		return
	}

	log.Info("Reading configuration file...")

	configDir := "ads/lab2/exercise2/files/configFiles/"
	node, err := loadNode(configDir + os.Args[1])
	if err != nil {
		log.Error("Error creating the node!", err)
		return
	}

	node.onIncoming()
}
