package main

import "net"
import "fmt"
import "os"
import "bufio"
//import "strings" // only needed below for sample processing

func reader(port string, stop chan bool){
	fmt.Println("Listening on port ", port)
	
	// listen on all interfaces
	ln, err := net.Listen("tcp", ":" + port)
  
	if err != nil {
	  fmt.Println("Error while starting server\n", err)
	  os.Exit(1)
	}

	conn, err := ln.Accept()

    if err != nil {
      fmt.Println("Error while accepting connection at server\n%v\n", err)
    } else {
      fmt.Println("Connection received at server")
    }

    for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			conn.Close()
			break
		}
		
		if string(message) == "stop\n" {			
			fmt.Println("Should stop\n") 
			conn.Close()
			stop <- true
		}

		// output message received
		fmt.Print("Message Received:", string(message))		
	}
}

func writer(otherIP, otherPort string, stop chan bool){
	// connect to this socket
	conn, err := net.Dial("tcp", otherIP + ":" + otherPort)
	for err != nil {
		conn, err = net.Dial("tcp", otherIP + ":" + otherPort)  //Protocol used, IP:port of server we're calling. Here conn is our channel.
	}	
	
	fmt.Println("Good dial")	 

	// run loop forever (or until ctrl-c)
	for { 
  
	  // read in input from stdin
	  reader := bufio.NewReader(os.Stdin)    
	  fmt.Print("Text to send: ")
	  text, _ := reader.ReadString('\n')
  
	  // send to socket
	  fmt.Fprintf(conn, text + "\n")   

	  if string(text) == "stop" {
		fmt.Println("Should stop\n") 
		stop <- true
	  }
	}
}

func main() {

  thisPort := "6001"
  otherPort := "6002"

  otherIP := "127.0.0.1"

  stop := make(chan bool)

  go reader(thisPort, stop);
  go writer(otherIP, otherPort, stop);

  <- stop
}
