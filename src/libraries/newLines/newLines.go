package newLines

import "strings" 

// normalizeNewLines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func NormalizeNewLines(inputText string) string {
	// replace CR LF \r\n (windows) with LF \n (unix)
	inputText = strings.Replace(inputText, "\r\n", "\n", -1)
	// replace CF \r (mac) with LF \n (unix)
	inputText = strings.Replace(inputText, "\r", "\n", -1)

	return inputText
}

// eraseNewLines erases \r\n (windows) and \r (mac) and \n (unix)
func EraseNewLines(inputText string) string {
	// replace CR LF \n (windows / linux) 
	inputText = strings.Replace(inputText, "\n", "", -1)
	// replace CF \r (mac)
	inputText = strings.Replace(inputText, "\r", "", -1)

	return inputText
}
