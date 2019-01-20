package main

import "net"
import "fmt"
import "os"
import "bufio"
import "strings" // only needed below for sample processing

func handleConnectionServer(c net.Conn) {

	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			c.Close()
			break
		}

		// output message received
		fmt.Print("Message Received:", string(message))

		// sample process for string received
		newmessage := strings.ToUpper(message)

		// send new string back to client
		c.Write([]byte(newmessage + "\n"))
	}

	fmt.Println("Connection closed at server")
}

func main() {

	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":6001")

	if err != nil {
		fmt.Println("Error while starting server\n", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Println("Error while accepting connection at server\n%v\n", err)
			continue
		} else {
			fmt.Println("Connection received at server")
		}

		go handleConnectionServer(conn)
	}

}
