package main

import (
	"bufio"
	"fmt"
	"io"
	"libraries/newLines"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func receiveFile(baseDir, filename string, connection net.Conn) error {

	receiveFileError := os.MkdirAll(baseDir, os.FileMode(int(0775)))
	if receiveFileError == nil {
		log.Info("READER - Directories ready!")

		var destinationFile *os.File
		destinationFile, receiveFileError = os.Create(baseDir + filename)
		if receiveFileError == nil {
			log.Info("READER - File ready!")
			defer destinationFile.Close()

			_, receiveFileError := io.Copy(destinationFile, connection)
			if receiveFileError == nil {
				log.Info("READER - Filename created at ", baseDir+filename)

				//Get file details
				file, _ := destinationFile.Stat()
				//EXC2. MOSTRAR EN PANTALLA NOMBRE Y TAMAÑO DEL ARCHIVO RECIBIDO
				log.Info("Name of file", file.Name())
				log.Info("Size of file", file.Size())

			} else {
				log.Error("READER - Error receiving file: ", receiveFileError)
			}

		} else {
			log.Error("READER - Error creating file: ", receiveFileError)
		}
	} else {
		log.Error("READER - Error creating directories: ", receiveFileError)
	}

	return receiveFileError
}

func readerLoop(connection net.Conn, stopChannel chan bool) {

	for {
		text, readError := bufio.NewReader(connection).ReadString('\n')
		if readError == nil {

			text = newLines.EraseNewLines(text)
			if text == "stop" {
				log.Info("READER - Stop received! Shutting down...")
				break

			} else {
				log.Info("READER - Filename received!")

				baseDir := "src/Lab1/exercise2/files/received/"
				receiveFileError := receiveFile(baseDir, text, connection)
				if receiveFileError != nil {
					log.Error("READER - Error receiving file: ", receiveFileError)
				}
			}

		} else {
			log.Error("READER - Error reading message: ", readError)
			if readError == io.EOF {
				log.Error("READER - EOF found, shuting down...")
				break
			}
		}
	}
}

func reader(port string, stopChannel chan bool) {

	log.Info("READER - Listening on port", port)

	// listen on all interfaces
	listener, listenerError := net.Listen("tcp", port)

	if listenerError == nil {
		connection, connectionError := listener.Accept()

		if connectionError == nil {
			log.Info("READER - Connection received at server")
			defer connection.Close()

			readerLoop(connection, stopChannel)
			stopChannel <- true

		} else {
			log.Error("READER - Error while accepting connection at server: ", connectionError)
		}

	} else {
		log.Error("READER - Error while starting server: ", listenerError)
	}
}

func sendFile(baseDir, filename string, connection net.Conn) error {

	//Open file to be sent
	file, sendFileError := os.Open(baseDir + filename)
	if sendFileError == nil {
		defer file.Close()

		//Send file details
		var fileDetails os.FileInfo
		fileDetails, sendFileError = file.Stat()
		if sendFileError == nil {

			_, sendFileError = fmt.Fprintf(connection, fileDetails.Name()+"\n")
			if sendFileError == nil {

				//Send file
				_, sendFileError = io.Copy(connection, file)
				// 	if sendFileError != nil {
				// 		log.Error("WRITER - Error sending file: ", sendFileError)
				// 	}

				// } else {
				// 	log.Error("WRITER - Error opening file: ", sendFileError)
			}

			// } else {
			// 	log.Error("WRITER - Error retrieving file details: ", sendFileError)
		}

		// } else {
		// 	log.Error("WRITER - Error opening file: ", sendFileError)
	}

	return sendFileError
}

func writerLoop(connection net.Conn, stopChannel chan bool) {

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("WRITER - Input full filename of file to send: ")
		text, readError := reader.ReadString('\n')

		if readError == nil {
			text = newLines.EraseNewLines(text)
			if "stop" == text {
				log.Info("WRITER - Should stop")
				fmt.Fprintf(connection, text+"\n") //TODO Handle error
				break
			} else if text == "" {
				// do nothing

			} else {
				baseDir := "src/Lab1/exercise2/files/"
				sendFileError := sendFile(baseDir, text, connection)
				if sendFileError != nil {
					log.Error("WRITER - Something went wrong while sending the file: ", sendFileError)
					cwd, _ := os.Getwd()
					log.Error("WRITER - CWD: ", cwd)
				}
			}
		} else {
			log.Error("WRITER - Error reading std input: ", readError)
		}
	}
}

func writer(otherIP, otherPort string, stopChannel chan bool) {

	// connect to this socket
	connection, connectionError := net.Dial("tcp", otherIP+otherPort)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", otherIP+otherPort)
	}
	defer connection.Close()

	log.Info("WRITER - Succesful dial")
	writerLoop(connection, stopChannel)
	stopChannel <- true
}

func main() {

	otherIP := "127.0.0.1"
	otherPort := ":6002"
	thisPort := ":6001"

	if len(os.Args) > 1 {
		log.Info("Ports set by arguments...")

		thisPort = os.Args[1]
		if !strings.HasPrefix(":", thisPort) {
			thisPort = ":" + thisPort
		}

		otherPort = os.Args[2]
		if !strings.HasPrefix(":", otherPort) {
			otherPort = ":" + otherPort
		}
	}

	stopChannel := make(chan bool)

	go reader(thisPort, stopChannel)
	go writer(otherIP, otherPort, stopChannel)

	<-stopChannel

}
