package lab2

import (
	"encoding/gob"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

// Node2 represents a node in a network, listening in his own port
// and sending messages to its neighbors
type Node2 struct {
	Address 	string			// Node Address
	Port        string          // Node listening port
	Neighbours  map[string]bool // Node neighbours
	ID          string
	Parent      string
	IsInitiator bool
}

// Message contain the information related to a received message.
type Message struct {
	OriginAddress string 	//	IP of sender
	OriginPort   string 	//	port of sender
	ID            string 	// 	ID of the sender
	Content       string 	// 	content of the message
	Error         error  	// 	contains the error, if any
}

func (node Node2) readMessage(connection net.Conn, incoming chan Message) {
	defer connection.Close()

	//Decode message received from connection
	messageDecoder := gob.NewDecoder(connection)
	message := Message{}
	decodeError := messageDecoder.Decode(&message)
	if decodeError != nil {
		log.Error(decodeError)
	} else {
		log.Debug("Message received from node ", message.ID + " - " + message.OriginAddress + message.OriginPort)
	}

	incoming <- message
}

// Listen establishes a connection to all the neighbours, returning
// channel to wait for incoming messages
func (node *Node2) Listen(incoming chan Message) error {
	log.Debug("Listen - Starting listener at port: ", node.Port)
	listener, listenerError := net.Listen("tcp", node.Port)
	if listenerError != nil {
		return listenerError
	}
	fmt.Println("Node is Online - Listening at port: ", node.Port)

	for {

		connection, connectionError := listener.Accept()
		if connectionError != nil {
			return connectionError
		}
		log.Debug("Listen - Connection received at server")

		go node.readMessage(connection, incoming)
	}
}

// SendMessage sends a message to a specific host
func (node *Node2) SendMessage(message Message, destination string, sent chan bool) {

	connection, connectionError := net.Dial("tcp", destination)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", destination)
	}
	defer connection.Close()
	log.Debug("SendMessage - Connected to ", destination)

	//Encode message and send to conenction
	messageEncoder := gob.NewEncoder(connection)
	messageToSend := message
	encodeError := messageEncoder.Encode(messageToSend)
	if encodeError != nil {
		log.Error(encodeError)
	} else {
		log.Debug("Message sent to ", destination)
	}


	if sent != nil {
		sent <- encodeError != nil
	}

}

// SendToAllNeighbours sends a message to all node's neighbours
func (node *Node2) SendToAllNeighbours(message Message, sent chan bool) {
	for neighbour := range node.Neighbours {
		go node.SendMessage(message, neighbour, sent)
	}
}
