package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"libraries/newLines"
)


func reader(port string, stopChannel chan bool) {
	fmt.Println("READER - Listening on port", port)

	// listen on all interfaces
	listener, listenerError := net.Listen("tcp", ":" + port)

	if listenerError != nil {
		fmt.Println("READER - Error while starting server\n", listenerError)
		os.Exit(1)
	}

	connection, connectionError := listener.Accept()

	if connectionError != nil {
		fmt.Println("READER - Error while accepting connection at server: \n", connectionError)
	} else {
		fmt.Println("READER - Connection received at server")
	}

	for {

		incomingFileName, readError := bufio.NewReader(connection).ReadString('\n')
		if readError != nil {
			connection.Close()
			break
		}

		destinationFile, fileCreationError := os.Create(incomingFileName)
		if fileCreationError != nil {
			fmt.Println("READER - Error creating file: ", fileCreationError)
		}

		_, fileReceptionError := io.Copy(destinationFile, connection)
		if fileReceptionError != nil {
			fmt.Println("READER - Error receiving file: ", fileReceptionError)
		}

		destinationFile.Close()

	}
}

func writer(otherIP, otherPort string, stopChannel chan bool) {
	// connect to this socket
	connection, connectionError := net.Dial("tcp", otherIP+":"+otherPort)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", otherIP+":"+otherPort)
	}
	fmt.Println("WRITER - Succesful dial")

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("WRITER - Input full filename of file to send: ")
		text, _ := reader.ReadString('\n')

		if "stop" == newLines.EraseNewLines(text) {
			fmt.Println("WRITER - Should stop")
			fmt.Fprintf(connection, text + "\n")		//TODO Handle error
			stopChannel <- true
		}

		//Open file to be sent
		file, openFileError := os.Open("/files/test.txt")
		if openFileError != nil {
			fmt.Println("WRITER - Error opening file: ", openFileError)
		}

		//Send file details
		fileDetails, fileDetailsError := file.Stat()
		if fileDetailsError != nil {
			fmt.Println("WRITER - Error retrieving file details: ", openFileError)
		}

		fmt.Fprintf(connection, fileDetails.Name() + "\n")		//TODO Handle error

		//Send file
		_, fileSendError := io.Copy(connection, file)
		if openFileError != nil {
			fmt.Println("WRITER - Error sending file: ", fileSendError)
		}

		file.Close()
	}
}

func main() {

	argsWithoutProg := os.Args[1:]

	thisPort := argsWithoutProg[0]
	otherPort := argsWithoutProg[1]

	otherIP := "127.0.0.1"

	stopChannel := make(chan bool)

	go reader(thisPort, stopChannel)
	go writer(otherIP, otherPort, stopChannel)

	<-stopChannel

}
