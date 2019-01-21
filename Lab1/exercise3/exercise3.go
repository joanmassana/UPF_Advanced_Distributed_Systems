package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"libraries/newLines"
	"net"
	"os"
	"strconv"
	"strings"
)

type networkConfig struct {
	server 		networkAddress
	clients 	[]networkAddress
}
type networkAddress struct {
	ip 		string
	port 	string
}

func receiveFile(baseDir, filename string, size int64, connection net.Conn) error {

	receiveFileError := os.MkdirAll(baseDir, os.FileMode(int(0775)))
	if receiveFileError == nil {
		log.Info("READER - Directories ready!")

		var destinationFile *os.File
		destinationFile, receiveFileError = os.Create(baseDir + filename)
		if receiveFileError == nil {
			log.Info("READER - File ready!")
			defer destinationFile.Close()

			_, receiveFileError := io.CopyN(destinationFile, connection, size)
			if receiveFileError == nil {
				log.Info("READER - Filename created at ", baseDir+filename)

				//Get file details
				file, _ := destinationFile.Stat()
				fmt.Println("Name of file: ", file.Name())
				fmt.Println("Size of file: ", file.Size())

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

func readMessage(connection net.Conn, stopChannel chan bool) {

	for {
		reader := bufio.NewReader(connection)
		data, readError := reader.ReadString('\n')
		log.Debug("Data received: ", data);
		if readError == nil {

			data = newLines.EraseNewLines(data)
			if data == "stop" {
				log.Info("READER - Stop received! Shutting down...")
				stopChannel <- true

			} else {
				filename := data
				log.Debug("READER - Filename received: ", filename)

				data, readError = reader.ReadString('\n')
				if readError == nil {

					data = newLines.EraseNewLines(data)

					var size int64
					size, readError = strconv.ParseInt(data, 10, 64)
					if readError == nil {
						log.Debug("READER - File size received: ", data)

						baseDir := "Lab1/exercise3/files/received/"
						receiveFileError := receiveFile(baseDir, filename, size, connection)
						if receiveFileError != nil {
							log.Error("READER - Error receiving file: ", receiveFileError)
						}
					} else {
						log.Error("READER - Error receiving file size: ", readError)
					}
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
		for {
			connection, connectionError := listener.Accept()

			if connectionError == nil {
				log.Info("READER - Connection received at server")

				go readMessage(connection, stopChannel)

			} else {
				log.Error("READER - Error while accepting connection at server: ", connectionError)
			}
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

			_, sendFileError = fmt.Fprintf(connection, "%s\n", fileDetails.Name())
			if sendFileError == nil {
				log.Debug("READER - Filename sent: ", fileDetails.Name())

				var bytesSent int
				bytesSent, sendFileError = fmt.Fprintf(connection, "%d\n", fileDetails.Size())
				if sendFileError == nil && bytesSent > 0 {
					log.Debug("READER - Filename size: ", fileDetails.Size())
					log.Debug("READER - Bytes sent: ", bytesSent)

					//Send file
					_, sendFileError = io.Copy(connection, file)
					// 	if sendFileError != nil {
					// 		log.Error("WRITER - Error sending file: ", sendFileError)
					// 	}

					// } else {
					// 	log.Error("WRITER - Error opening file: ", sendFileError)
				}
			}

			// } else {
			// 	log.Error("WRITER - Error retrieving file details: ", sendFileError)
		}

		// } else {
		// 	log.Error("WRITER - Error opening file: ", sendFileError)
	}

	return sendFileError
}

func getUserInput(networkConfiguration networkConfig, stopChannel chan bool) {

	//Establish connections with other nodes
	var connections []net.Conn
	for _, client := range networkConfiguration.clients {
		newConnection := dial(client.ip, client.port)
		connections = append(connections, newConnection)
	}

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Input full filename of file to send: ")
		text, readError := reader.ReadString('\n')

		if readError == nil {
			text = newLines.EraseNewLines(text)
			if "stop" == text {
				log.Info("WRITER - Should stop")
				for _, connection := range connections {
					fmt.Fprintf(connection, text+"\n") //TODO Handle error
				}
				stopChannel <- true
			} else if text == "" {
				// do nothing

			} else {
				baseDir := "Lab1/exercise3/files/"
				for _, connection := range connections {
					log.Debug("Connection:: ", connection)
					sendFileError := sendFile(baseDir, text, connection)
					if sendFileError != nil {
						log.Error("WRITER - Something went wrong while sending the file: ", sendFileError)
						cwd, _ := os.Getwd()
						log.Error("WRITER - CWD: ", cwd)
					} else {
						log.Info("File sent")
					}
				}
			}
		} else {
			log.Error("WRITER - Error reading std input: ", readError)
		}
	}
}

func dial(otherIP, otherPort string) net.Conn{

	// connect to this socket
	connection, connectionError := net.Dial("tcp", otherIP+otherPort)
	for connectionError != nil {
		connection, connectionError = net.Dial("tcp", otherIP+otherPort)
	}

	log.Info("WRITER - Succesful dial - Connected to ", otherIP+otherPort)

	return connection

}

func getNetworkConfig(configFilePath string) networkConfig {

	var config networkConfig

	addresses, _ := fileToLines(configFilePath)

	config.server = getNetworkAddress(addresses[0])
	for i := 1; i < len(addresses); i++ {
		client := getNetworkAddress(addresses[i])
		config.clients = append(config.clients, client)
	}

	log.Info("Server: ", config.server)
	log.Info("Clients: ", config.clients)

	return config
}

func fileToLines(filePath string) (lines []string, err error) {
	f, err := os.Open("Lab1/exercise3/files/configFiles/" + filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	return
}

func getNetworkAddress(address string) networkAddress{
	var networkAddress networkAddress

	addressSlice := strings.Split(address, ":")
	networkAddress.ip = addressSlice[0]
	networkAddress.port = ":" + addressSlice[1]

	return networkAddress
}

func main() {

	log.SetLevel(log.DebugLevel)
	var networkConfiguration networkConfig
	stopChannel := make(chan bool)

	if len(os.Args) > 0 {
		//Get config file as argument
		configFilePath := os.Args[1]

		//get networkConfig struct
		networkConfiguration = getNetworkConfig(configFilePath)

	} else {
		log.Error("Argument missing: Config file.")
	}

	go reader(networkConfiguration.server.port,stopChannel)
	go getUserInput(networkConfiguration, stopChannel)

	<-stopChannel

}
