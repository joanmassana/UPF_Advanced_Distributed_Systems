package main

import "net"
import "fmt"
import "os"
import "bufio"
import "libraries/newLines"

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
		// will listen for message to process ending in newline (\n)
		message, readError := bufio.NewReader(connection).ReadString('\n')
		if readError != nil {
			connection.Close()
			break
		}

		// output message received
		fmt.Print("READER - Message Received: ", string(message))

		if "stop" == newLines.EraseNewLines(message) {
			fmt.Println("READER - Should stop")
			connection.Close()
			stopChannel <- true
		}
	}
}

func writer(otherIP, otherPort string, stopChannel chan bool) {
	// connect to this socket
	connection, connectionError := net.Dial("tcp", otherIP+":"+otherPort)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", otherIP+":"+otherPort)
	}
	fmt.Println("WRITER - Succesful dial")

	// run loop forever (or until ctrl-c)
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("WRITER - Text to send: ")
		text, _ := reader.ReadString('\n')

		// send to socket if there is something to send
		if len(newLines.EraseNewLines(text)) > 0 {
			fmt.Fprintf(connection, text+"\n")
		}

		if "stop" == newLines.EraseNewLines(text) {
			fmt.Println("WRITER - Should stop")
			stopChannel <- true
			//TODO Aqui hay un caso en que si el mensaje a stopChannel tarda en llegar, vemos otra vez el 'Text to send' en consola
		}
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
