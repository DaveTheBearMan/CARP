package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Map to store connections by IP address
var clients = make(map[string]*websocket.Conn)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Handle upgrading out client from an http request to a websocket request.
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer connection.Close()

	// Extract client IP address for storing them in a table
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("Error extracting IP:", err)
		return
	}

	// Store the WebSocket connection by client IP
	clients[clientIP] = connection
	log.Printf("Client connected: %s", clientIP)

	// Keep the connection open for sending commands
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			delete(clients, clientIP)
			break
		}

		log.Printf("Received response from client [%s]: %s", clientIP, message)
	}
}

func sendCommandToClients(command string) {
	fullCommand := strings.TrimSpace(command)
	for clientIP, clientConn := range clients {
		if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
			log.Printf("Write error for client [%s]: %s", clientIP, err)
			clientConn.Close()
			delete(clients, clientIP)
		} else {
			log.Printf("Command sent to client [%s]: %s", clientIP, fullCommand)
		}
	}
}

func sendCommandToClient(ip_addr string, command string) {
	fullCommand := strings.TrimSpace(command)
	clientConn := clients[ip_addr]

	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
		log.Printf("Write error for client [%s]: %s", ip_addr, err)
		clientConn.Close()
		delete(clients, ip_addr)
	} else {
		log.Printf("Command sent to client [%s]: %s", ip_addr, fullCommand)
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	go func() {
		// Simulate sending commands to clients
		for {
			reader := bufio.NewReader(os.Stdin)

			// Get ip address
			fmt.Print("Enter ip address: ")
			client_addr, _ := reader.ReadString('\n')
			client_addr = client_addr[:len(client_addr)-1]

			// Get command
			fmt.Print("Enter command: ")
			input, _ := reader.ReadString('\n')
			// Remove the trailing newline
			input = input[:len(input)-1]

			sendCommandToClient(client_addr, input)
			//sendCommandToClients(input)
		}
	}()

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
