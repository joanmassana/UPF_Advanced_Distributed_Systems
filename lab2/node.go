package lab2

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

// Node represents a node in a network, listening in his own port
// and sending messages to its neighbors
type Node struct {
	Port        string   // Node listening port
	Neighbours  map[string]bool // Node neighbours
	StopChannel chan bool
	Id string
	Parent string
	IsInitiator bool
	// Connnections to its neighbours
}

func (node *Node) readMessage(connection net.Conn, stopChannel chan bool) error {
	defer connection.Close()

	reader := bufio.NewReader(connection)

	data, readError := reader.ReadString('\n')
	if readError != nil {
		return readError
	}
	log.Debug("READER - Raw data received: ", data)


	text := EraseNewLines(data)
	if text == "stop" {
		log.Info("READER - Stop received! Shutting down...")
		stopChannel <- true

	} else {
		fmt.Println(text)
		node.updateReceivedMap(text)

		if hasReceivedAllNeighbours(node) {
			//sendToParent
		} else {
			stopSent := make(chan bool)
			for neighbour := range node.Neighbours {
				go connectAndSend(neighbour, "127.0.0.1" + node.Port + ":" + node.Id, stopSent)
			}
		}

	}

	return nil
}

func (node *Node) updateReceivedMap(message string){
	slice := strings.Split(message, ":")
	host := slice[0] + slice[1]
	node.Neighbours[host] = true

	if node.Parent == "" {
		node.Parent = host
	}
}

func hasReceivedAllNeighbours(node *Node) bool{
	for _, value := range node.Neighbours {
		if !value {
			return false;
		}
	}
	return true;
}

func (node *Node) listenNeighbours() error {

	log.Info("READER - Starting listener at port: ", node.Port)
	listener, listenerError := net.Listen("tcp", node.Port)
	if listenerError != nil {
		return listenerError
	}
	log.Info("READER - Listening...")

	for {

		connection, connectionError := listener.Accept()
		if connectionError != nil {
			return connectionError
		}
		log.Info("READER - Connection received at server")

		go node.readMessage(connection, node.StopChannel)
	}
}

func sendMessage(message string, connection net.Conn) error {
	_, sendMessageError := fmt.Fprintf(connection, "%s\n", message)
	if sendMessageError != nil {
		return sendMessageError
	}
	log.Debug("WRITER - Filename sent: ", message)

	return nil
}


func connectAndSend(neighbour, message string, stopSent chan bool) error {

	log.Info("WRITER - trying to connect to ", neighbour)
	connection, connectionError := net.Dial("tcp", neighbour)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", neighbour)
	}
	defer connection.Close()
	log.Info("WRITER - Connected to ", neighbour)

	if "stop" == message {
		log.Info("WRITER - Should stop")

		_, writeConnError := fmt.Fprintf(connection, message+"\n")
		stopSent <- true
		if writeConnError != nil {
			log.Error("WRITER - Something went wrong while sending the stop: ", writeConnError)
			return writeConnError
		}

	} else {
		sendMessageError := sendMessage(message, connection)
		if sendMessageError != nil {
			log.Error("WRITER - Something went wrong while sending the stop: ", sendMessageError)
			return sendMessageError
		}
	}
	return nil
}

func (node *Node) waitUserInput() {

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Input full filename of file to send: ")
		data, readError := reader.ReadString('\n')

		if readError != nil {
			log.Error("WRITER - Error reading std input: ", readError)
		}

		text := EraseNewLines(data)
		log.Info("WRITER - Text to send: '", text, "'")

		if text == "" {
			// do nothing
			continue
		}

		stopSent := make(chan bool)
		for _, neighbour := range node.Neighbours {
			go connectAndSend(neighbour, text, stopSent)
		}

		if text == "stop" {
			for range node.Neighbours {
				log.Debug("WRITER - Waiting stop sent...")
				<-stopSent
				log.Debug("WRITER - Stop sent!")
			}
			break
		}
	}
	node.StopChannel <- true
}

// Start the node, which listen for neighbours and waits for user input.
func (node *Node) Start() {
	go node.listenNeighbours()
	go node.waitUserInput()
}

func main() {

	stopChannel := make(chan bool)

	node := Node{
		Port: ":6001",
		Neighbours: []string{
			"127.0.0.1:6002",
			"127.0.0.1:6003",
		},
		StopChannel: stopChannel,
	}
	go node.Start()

	<-stopChannel
}
