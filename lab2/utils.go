package lab2

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

/**
FILE UTILS
 */
func SendFile(filepath string, connection net.Conn) error {

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


/**
NEWLINES UTILS
 */
// NormalizeNewLines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func NormalizeNewLines(inputText string) string {
	// replace CR LF \r\n (windows) with LF \n (unix)
	inputText = strings.Replace(inputText, "\r\n", "\n", -1)
	// replace CF \r (mac) with LF \n (unix)
	inputText = strings.Replace(inputText, "\r", "\n", -1)

	return inputText
}

// EraseNewLines erases \r\n (windows) and \r (mac) and \n (unix)
func EraseNewLines(inputText string) string {
	// replace CR LF \n (windows / linux)
	inputText = strings.Replace(inputText, "\n", "", -1)
	// replace CF \r (mac)
	inputText = strings.Replace(inputText, "\r", "", -1)

	return inputText
}
