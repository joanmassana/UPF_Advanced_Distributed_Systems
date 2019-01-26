package lab1

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// Node represents a node in a network, listening in his own port
// and sending messages to its neighbors
type Node struct {
	Port        string    // Node listening port
	Neighbours  []string  // Node neighbours
	StopChannel chan bool // Connnections to its neighbours
}

func receiveFile(baseDir, filename string, connection net.Conn, size int64) error {

	mkdirsError := os.MkdirAll(baseDir, os.FileMode(int(0775)))
	if mkdirsError != nil {
		return mkdirsError
	}
	log.Info("READER - Directories ready!")

	filepath := baseDir + filename
	destinationFile, createFileError := os.Create(filepath)
	if createFileError != nil {
		return createFileError
	}
	defer destinationFile.Close()
	log.Info("READER - File ready!")

	_, receiveFileError := io.CopyN(destinationFile, connection, size)
	if receiveFileError != nil {
		return receiveFileError
	}
	log.Info("READER - File contents stored at ", filepath)

	return nil
}

func readMessage(connection net.Conn, stopChannel chan bool) error {
	defer connection.Close()

	reader := bufio.NewReader(connection)

	data, readError := reader.ReadString('\n')
	if readError != nil {
		return readError
	}
	log.Debug("READER - Data received: ", data)

	text := EraseNewLines(data)
	if text == "stop" {
		log.Info("READER - Stop received! Shutting down...")
		stopChannel <- true

	} else {
		filename := text
		log.Info("READER - File name: ", filename)

		data, readError = reader.ReadString('\n')
		if readError != nil {
			return readError
		}
		log.Debug("READER - Data received: ", data)

		size, parseError := strconv.ParseInt(data, 10, 64)
		if parseError != nil {
			return parseError
		}
		log.Info("READER - File size: ", size)

		fmt.Println("READER - Name of file: ", filename)
		fmt.Println("READER - Size of file: ", size)

		//receiveFile(filename, size)
	}

	return nil
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

		go readMessage(connection, node.StopChannel)
	}
}

func sendFile(filepath string, connection net.Conn) error {

	//Open file to be sent
	file, openFileError := os.Open(filepath)
	if openFileError != nil {
		return openFileError
	}
	defer file.Close()

	//Send file details
	fileDetails, statFileError := file.Stat()
	if statFileError != nil {
		return statFileError
	}

	_, sendFilenameError := fmt.Fprintf(connection, "%s\n", fileDetails.Name())
	if sendFilenameError != nil {
		return sendFilenameError
	}
	log.Debug("WRITER - Filename sent: ", fileDetails.Name())

	bytesSent, sendFileSizeError := fmt.Fprintf(connection, "%d\n", fileDetails.Size())
	if sendFileSizeError != nil {
		return sendFileSizeError
	}
	log.Debug("WRITER - Filename size: ", fileDetails.Size())
	log.Debug("WRITER - Bytes sent: ", bytesSent)

	/*
		_, copyFileError := io.Copy(connection, file)
		if copyFileError != nil {
			return copyFileError
		}
	*/
	return nil
}

func handleMessage(neighbour, message string, stopSent chan bool) error {

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
		sendFileError := sendFile(message, connection)
		if sendFileError != nil {
			log.Error("WRITER - Something went wrong while sending the file: ", sendFileError)
			return sendFileError
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
			go handleMessage(neighbour, text, stopSent)
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
