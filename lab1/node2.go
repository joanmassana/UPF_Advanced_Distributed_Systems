package lab1

import (
	"bufio"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

// Node represents a node in a network, listening in his own port
// and sending messages to its neighbors
type Node2 struct {
	Port        string    // Node listening port
	Neighbours  []string  // Node neighbours
	StopChannel chan bool // Connnections to its neighbours
}

// Message contain the information related to a received message.
type Message struct {
	ID      string // ID of the sender
	Content string // content of the message
	Error   error  // contains the error, if any
}

func (node Node2) readMessage(connection net.Conn, incoming chan Message) {
	defer connection.Close()

	message := Message{
		ID:      connection.RemoteAddr().String(),
		Content: "",
		Error:   nil,
	}
	reader := bufio.NewReader(connection)

	data, readError := reader.ReadString('\n')
	if readError != nil {
		message.Error = readError
	} else {
		log.Debug("readMessage - Data received: ", data)

		text := EraseNewLines(data)
		message.Content = text
	}
	incoming <- message
}

// Listen stablishes a connection to all the neighbours, returning
// channel to wait for incoming messages
func (node *Node2) Listen(incoming chan Message) error {

	log.Info("Listen - Starting listener at port: ", node.Port)
	listener, listenerError := net.Listen("tcp", node.Port)
	if listenerError != nil {
		return listenerError
	}
	log.Info("Listen - Listening...")

	for {

		connection, connectionError := listener.Accept()
		if connectionError != nil {
			return connectionError
		}
		log.Info("Listen - Connection received at server")

		go node.readMessage(connection, incoming)
	}
}

// SendMessage sends a message to a specific host
func (node *Node2) SendMessage(text string, host string, sent chan bool) {

	connection, connectionError := net.Dial("tcp", host)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", host)
	}
	defer connection.Close()
	log.Info("SendMessage - Connected to ", host)

	_, writeConnError := fmt.Fprintf(connection, text+"\n")
	if writeConnError != nil {
		log.Error("SendMessage - Something went wrong while sending a message: ", writeConnError)
	}

	if sent != nil {
		sent <- writeConnError != nil
	}
}

// SendToAllNeighbours sends a message to all node's neighbours
func (node *Node2) SendToAllNeighbours(text string, sent chan bool) {
	for _, neighbour := range node.Neighbours {
		go node.SendMessage(text, neighbour, sent)
	}
}
