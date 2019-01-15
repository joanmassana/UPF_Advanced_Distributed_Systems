package main

import "net"
import "fmt"
import "os"
import "bufio"
import "strings" // only needed below for sample processing

func reader(port string, stopChannel chan bool){
	fmt.Println("Listening on port ", port)
	
	// listen on all interfaces
	ln, err := net.Listen("tcp", ":" + port)
  
	if err != nil {
	  fmt.Println("Error while starting server\n", err)
	  os.Exit(1)
	}

	conn, err := ln.Accept()

    if err != nil {
      fmt.Println("Error while accepting connection at server: \n", err)
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
		
		// output message received
		fmt.Print("Message Received: ", string(message))	
		
		if "stop" == eraseNewlines(message) {
			fmt.Println("Reader - Should stop") 
			conn.Close()
			stopChannel <- true
		}
			
	}
}

func writer(otherIP, otherPort string, stopChannel chan bool){
	// connect to this socket
	conn, err := net.Dial("tcp", otherIP + ":" + otherPort)
	for err != nil {
		conn, err = net.Dial("tcp", otherIP + ":" + otherPort)  //Protocol used, IP:port of server we're calling. Here conn is our channel.
	}	
	
	fmt.Println("Succesful dial")	 

	// run loop forever (or until ctrl-c)
	for { 
  
	  // read in input from stdin
	  reader := bufio.NewReader(os.Stdin)    
	  fmt.Print("Text to send: ")
	  text, _ := reader.ReadString('\n')
  
	  // send to socket
	  fmt.Fprintf(conn, text + "\n")   

	  if "stop" == eraseNewlines(text) {
			fmt.Println("Writer - Should stop")
			stopChannel <- true
			//TODO Aqui hay un caso en que si el mensaje a stopChannel tarda en llegar, vemos otra vez el 'Text to send' en consola
	  }
	}
}

// normalizeNewLines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func normalizeNewlines(inputText string) string {
	// replace CR LF \r\n (windows) with LF \n (unix)
	inputText = strings.Replace(inputText, "\r\n", "\n", -1)
	// replace CF \r (mac) with LF \n (unix)
	inputText = strings.Replace(inputText, "\r", "\n", -1)

	return inputText
}

// eraseNewLines erases \r\n (windows) and \r (mac) and \n (unix)
func eraseNewlines(inputText string) string {
	// replace CR LF \n (windows / linux) 
	inputText = strings.Replace(inputText, "\n", "", -1)
	// replace CF \r (mac)
	inputText = strings.Replace(inputText, "\r", "", -1)

	return inputText
}

func main() {

  thisPort := "6001"
  otherPort := "6002"

  otherIP := "127.0.0.1"

  stopChannel := make(chan bool)

  go reader(thisPort, stopChannel);
  go writer(otherIP, otherPort, stopChannel);

  <- stopChannel
}
