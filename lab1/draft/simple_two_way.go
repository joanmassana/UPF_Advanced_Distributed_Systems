package main

import "net"
import "fmt"
import "os"
import "bufio"

//import "strings" // only needed below for sample processing

func listenAndOutput(port string, stop chan bool) {

	// listen on all interfaces
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error while starting two way client\n", err)
		os.Exit(1)
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error while accepting connection at two way client", err)
	} else {
		fmt.Println("Connection received at two way client")
	}

	reader := bufio.NewReader(conn)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			break
		}

		// output message received
		if message == "stop\n" {
			fmt.Println("Stop message received!")
			stop <- true
		} else {
			fmt.Print("\n<>Message Received:", string(message))
			fmt.Print(">")
		}
	}
	fmt.Println("Connection closed")
}

func dialAndInput(ip, port string, stop chan bool) {

	conn, err := net.Dial("tcp", ip+port)
	for err != nil {
		conn, err = net.Dial("tcp", ip+port)
	}

	// read in input from stdin
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		text, _ := reader.ReadString('\n')

		// send to socket
		fmt.Fprintf(conn, text)
		if text == "stop\n" {
			fmt.Println("Stop message sent!")
			stop <- true
		}
	}
}

func main() {

	otherIP := "127.0.0.1"
	otherPort := ":6002"
	thisPort := ":6001"
	if len(os.Args) > 1 {
		fmt.Println("Ports set by arguments...")
		thisPort = os.Args[1]
		otherPort = os.Args[2]
	}

	stop := make(chan bool)

	go listenAndOutput(thisPort, stop)
	go dialAndInput(otherIP, otherPort, stop)

	<-stop
}
