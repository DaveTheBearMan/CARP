package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

// logEntry struct to hold the destination and message
type logEntry struct {
	location *tview.TextView // The destination TextView
	message  string          // The message to log
}

// Function to handle upgrading to websocket from HTTP requests
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Map of all client websockets
var clients = make(map[string]*websocket.Conn)
var clientUser = make(map[string]string)

// String references
var targetClient string

// tview Application and children
var app *tview.Application
var selectedClient *tview.TextView
var activeClients *tview.TextView
var outputLog *tview.TextView
var connectionLog *tview.TextView
var errorLog *tview.TextView

// Buffered channel for logs
var logChan = make(chan logEntry, 100)

// Takes in a map containing IP address indices and websocket connections, and returns all available IPs.
func writeKeysFromMap(m map[string]*websocket.Conn, prefix string) string {
	b := new(bytes.Buffer)
	for key := range m {
		fmt.Fprintf(b, "%s%s\n", prefix, key)
	}
	return b.String()
}

// Set up logging goroutine.
func init() {
	go func() {
		// Loop over each entry in the log channel, and attempt to write to that buffer.
		for entry := range logChan {
			if _, err := fmt.Fprint(entry.location, entry.message); err != nil {
				log.Printf("Error writing to location: %v", err)
				queueMessageToBeWritten(errorLog, "Error writing to location: %v", err)
			}
			entry.location.ScrollToEnd()
			// Redraw the app to ensure updates come across.
			app.Draw()
		}
	}()
}

// Take in a place to update and a string/format, then run app.Draw on it
func queueMessageToBeWritten(location *tview.TextView, baseString string, args ...any) {
	message := fmt.Sprintf(baseString, args...)
	logChan <- logEntry{location: location, message: message}
}

// Write out our current user, directory, and host
func callingInformation(clientIP string) {
	// Check if the index exists, and if not, create it. Avoid rerunning commands for hosts
	caller, ok := clientUser[clientIP]
	if !ok {
		// Get user, host
		user, err := readCommandFromClient(clientIP, "whoami")
		if err != nil {
			queueMessageToBeWritten(outputLog, "%s", "Unable to get whoami")
			return
		}
		user = strings.ReplaceAll(user, "\n", "")
		user = strings.ReplaceAll(user, "\t", "")
		user = strings.ReplaceAll(user, " ", "")
		hostname, err := readCommandFromClient(clientIP, "hostname")
		if err != nil {
			queueMessageToBeWritten(outputLog, "%s", "Unable to get hostname")
			return
		}
		hostname = strings.ReplaceAll(hostname, "\n", "")
		hostname = strings.ReplaceAll(hostname, "\t", "")
		hostname = strings.ReplaceAll(hostname, " ", "")
		
		clientUser[clientIP] = fmt.Sprintf("[green]%s@%s[-]: $ ", user, hostname)
	} else {
		// Write to the output buffer the caller
		queueMessageToBeWritten(outputLog, caller)
	}
}

// Handle upgrading out client from an http request to a websocket request.
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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
		queueMessageToBeWritten(outputLog, "Error extracting IP: %s", err)
		return
	}

	// Private IP address creation from Hostname -I
	privateIP, err := readCommandFromSocket(connection, clientIP, "hostname -I | awk '{print $1}'")
	if err != nil {
		queueMessageToBeWritten(outputLog, "Error extracting private IP: %s", err)
		return
	}
	privateIP = strings.TrimSpace(privateIP)

	// Add client connection I guess.
	clients[privateIP] = connection
	callingInformation(privateIP)

	// Clear active clients, and write our current table.
	activeClients.Clear()
	queueMessageToBeWritten(activeClients, "%s", writeKeysFromMap(clients, ""))

	// Log a client connected to our connection log
	queueMessageToBeWritten(connectionLog, "[[green]+[-]] [yellow]%s\n[-]", privateIP)

	// Keep the connection open for sending commands
	//callingInformation(privateIP)
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			//log.Println("Read error:", err)
			queueMessageToBeWritten(outputLog, "Read error: %s\n", err)
			queueMessageToBeWritten(connectionLog, "[[red]-[-]] [yellow]%s\n[-]", privateIP)
			delete(clients, privateIP)
			delete(clientUser, privateIP)
			// Clear active clients, and write our current table. Removes the fragmented IP
			activeClients.Clear()
			queueMessageToBeWritten(activeClients, "%s", writeKeysFromMap(clients, ""))
			break
		}

		// Send command to all clients or a specified target.
		if targetClient == "all" {
			// Flatten responses to one line for each client, and include their IP address.
			message = []byte(strings.Replace(string(message), "\n", " ", -1))
			queueMessageToBeWritten(outputLog, "[yellow]%s[-]\t", privateIP)
			callingInformation(privateIP)
			queueMessageToBeWritten(outputLog, "%s\n", message)
		} else {
			queueMessageToBeWritten(outputLog, "%s\n", message)
		}
	}
}

// Sends a command to a particular client, and saves the command for later printing.
func sendCommandToClient(ip_addr string, command string) {
	// Take in a command and ip address, and try across the client array for connecting by IP.
	fullCommand := strings.TrimSpace(command)
	clientConn := clients[ip_addr]

	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
		queueMessageToBeWritten(errorLog, "Write error for client: %s\n", err)
		clientConn.Close()
		delete(clients, ip_addr)
		return
	}

	// Update last command so we can print it later on for better record keeping.
	callingInformation(ip_addr)
	queueMessageToBeWritten(outputLog, "%s", command)
}

// Read a command from a specified socket, and return the response.
func readCommandFromSocket(clientConn *websocket.Conn, ip_addr string, command string) (string, error) {
	fullCommand := strings.TrimSpace(command)

	// Send the command to the client
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
		return "", fmt.Errorf("write error for client %s: %v", ip_addr, err)
	}

	// Read the response from the client
	_, response, err := clientConn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("read error for client %s: %v", ip_addr, err)
	}

	// Return the response as a string
	return string(response), nil
}

// Send a comand to all attached clients
func sendCommandToClients(command string) {
	fullCommand := strings.TrimSpace(command)
	for clientIP, clientConn := range clients {
		if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
			queueMessageToBeWritten(errorLog, "Write error for client [%s]: %s", clientIP, err)
			clientConn.Close()
			delete(clients, clientIP)
		}
	}
	queueMessageToBeWritten(outputLog, "[green]Command sent to clients [-][[blue] %s [-]]: %s\n", targetClient, fullCommand)
}

// Read a command from the client rather than print it.
func readCommandFromClient(ip_addr string, command string) (string, error) {
	fullCommand := strings.TrimSpace(command)
	clientConn := clients[ip_addr]

	// Send the command to the client
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {

		return "", fmt.Errorf("write error for client %s: %v", ip_addr, err)
	}

	// Read the response from the client
	_, response, err := clientConn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("read error for client %s: %v", ip_addr, err)
	}

	// Return the response as a string
	return string(response), nil
}

// Main function
func main() {
	// Open new application abd web server for web sockets
	http.HandleFunc("/ws", handleWebSocket)
	app = tview.NewApplication()

	// Selected Client
	selectedClient = tview.NewTextView()
	selectedClient.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Active Client ")

	// Active clients
	activeClients = tview.NewTextView()
	activeClients.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Clients ")

	// Connection log
	connectionLog = tview.NewTextView()
	connectionLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Connection Log ")

	// Error log
	errorLog = tview.NewTextView()
	errorLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Error Log ")

	// Input field for command name
	var commandField *tview.InputField
	commandField = tview.NewInputField().
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				command := commandField.GetText()
				if targetClient == "all" {
					if command == "clear" {
						go app.QueueUpdateDraw(func() {
							outputLog.Clear()
							commandField.SetText("")
						})
					} else {
						commandField.SetText("")
						sendCommandToClients(command)
					}
				} else if targetClient != "" {
					if command == "clear" {
						go app.QueueUpdateDraw(func() {
							outputLog.Clear()
							commandField.SetText("")
						})
					} else {
						commandField.SetText("")
						sendCommandToClient(targetClient, command)
					}
				}
			}
		})
	commandField.SetTitle(" Send Command ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Input field for client name
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel(" Client IP: ").
		SetLabelColor(tcell.ColorWhite).
		SetAcceptanceFunc(tview.InputFieldMaxLength(20)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				ip_addr := inputField.GetText()
				if ip_addr == "all" {
					targetClient = "all"
					inputField.SetText("")
					outputLog.Clear()
					selectedClient.SetText(" All clients ")
					return
				}

				_, ok := clients[ip_addr]
				if ok {
					targetClient = ip_addr
					selectedClient.SetText(ip_addr)
					inputField.SetText("")
					outputLog.Clear()
					go callingInformation(ip_addr)
				} else {
					inputField.SetText("Invalid IP address!")
					time.Sleep(2 * time.Second)
					inputField.SetText("")
				}
			}
		})

	inputField.SetTitle(" Select Client ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Text view to display the list of clients
	outputLog = tview.NewTextView()
	outputLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Response ")

	// Assemble the layout
	flex := tview.NewFlex().
		AddItem(tview.NewFlex().
			AddItem(inputField, 0, 1, false).
			AddItem(selectedClient, 0, 1, false).
			AddItem(activeClients, 0, 10, false).SetDirection(tview.FlexRow), 0, 2, true).
		AddItem(tview.NewFlex().
			AddItem(outputLog, 0, 10, false).
			AddItem(commandField, 0, 1, false).SetDirection(tview.FlexRow), 0, 5, false).
		AddItem(tview.NewFlex().
			AddItem(connectionLog, 0, 1, false).
			AddItem(errorLog, 0, 1, false).SetDirection(tview.FlexRow), 0, 2, false)

	// Start the HTTP server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
	//outputLog.SetText("Server started on http://localhost:8080\n")
}
