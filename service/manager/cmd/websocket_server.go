package cmd

import (
	"fmt"
	"http_proxy/manager/client"
	"http_proxy/manager/ui"
	"http_proxy/manager/utils"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

// Make sure that we have last command
var lastCommand string

// Function to handle upgrading to websocket from HTTP requests
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handle command
func RunCommand(commandField *tview.InputField) {
	command := commandField.GetText()
	if command == "clear" {
		go ui.App.QueueUpdateDraw(func() {
			ui.OutputLog.Clear()
			commandField.SetText("")
		})
		return
	}

	// Determine whether we need a bulk request (all,wildcard,alias) or a singular request (unique ip)
	switch client.TargetClient {
	case "all":
		BulkRequest(command, client.Clients)
	case "wildcard":
		BulkRequest(command, client.WildcardArray)
	case "alias":
		BulkRequest(command, client.AliasArray)
	default:
		commandField.SetText("")
		PostCommand(client.TargetClient, command)
	}
	commandField.SetText("")
}

// Handle upgrading out client from an http request to a websocket request.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer connection.Close()

	// Extract client IP address for storing them in a table
	incoming_ip_addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Error extracting IP: %s", err)
		return
	}

	// Get private IP address
	private_ip_addr, err := postCommandAndListenInternalSilenced(connection, incoming_ip_addr, "CARP-ip")
	if err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Error extracting private IP from CARP-ip: %s", err)
		return
	}

	// Hostname
	hostname, err := postCommandAndListenInternalSilenced(connection, incoming_ip_addr, "CARP-hostname")
	if err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Error extracting hostname from CARP-hostname: %s", err)
		return
	}

	// Add client connection
	client.Clients[private_ip_addr] = connection
	client.ClientUser[private_ip_addr] = hostname
	defer ui.QueueClientRemoval(private_ip_addr)

	// Clear active clients, and write our current table.
	ui.ActiveClients.Clear()
	ui.QueueMessageToBeWritten(ui.ActiveClients, "%s", utils.WriteKeysFromMap(""))

	// Log a client connected to our connection log
	ui.QueueMessageToBeWritten(ui.ConnectionLog, "[[green]+[-]] [yellow]%s\n[-]", private_ip_addr)

	// Keep the connection open for sending commands
	//callingInformation(privateIP)
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			ui.QueueMessageToBeWritten(ui.ErrorLog, "Read error: %s\n", err)
			break
		}

		// Send command to all clients or a specified target.
		if client.TargetClient == "all" || client.TargetClient == "wildcard" || client.TargetClient == "alias" {
			// Flatten responses to one line for each client, and include their IP address.
			message = []byte(strings.Replace(string(message), "\n", " ", -1))
			ui.QueueMessageToBeWritten(ui.OutputLog, "[yellow]%-*s[-]%s %s\n", 18, private_ip_addr, client.ClientUser[private_ip_addr], message)
		} else {
			switch lastCommand {
			case "cd":
				client.ClientUser[private_ip_addr] = string(message)
			default:
				ui.QueueMessageToBeWritten(ui.OutputLog, "%s\n", message)
			}
		}
	}
}

func PostCommand(ip_addr string, command string) {
	// Take in a command and ip address, and try across the client array for connecting by IP.
	trimmedCommand := strings.TrimSpace(command)
	clientConn := client.Clients[ip_addr]
	clientTitle := client.ClientUser[ip_addr]
	lastCommand = strings.Fields(trimmedCommand)[0]

	// Post command
	// This will send the command across the web socket, and catch any errors that are sent back in response.
	// Theres no need to print the full response here, as the websocket will do that automatically in the main
	// loop.
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(trimmedCommand)); err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Write error for client: %s\n", err)
		clientConn.Close()
		return
	}

	ui.QueueMessageToBeWritten(ui.OutputLog, "%s %s\n", clientTitle, trimmedCommand)
}

// Wraps the post command function to take in just an ip address, rather than the list as well. Boiler plate
func PostCommandAndListen(ip_addr string, command string) (string, error) {
	clientConn := client.Clients[ip_addr]

	return postCommandAndListenInternal(clientConn, ip_addr, command)
}

// TODO: Refactor into one command
func postCommandAndListenInternalSilenced(clientConn *websocket.Conn, ip_addr string, command string) (string, error) {
	// Take in a command and ip address, and try across the client array for connecting by IP.
	trimmedCommand := strings.TrimSpace(command)

	// Post command
	// This will send the command across the web socket, and catch any errors that are sent back in response.
	// Theres no need to print the full response here, as the websocket will do that automatically in the main
	// loop.
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(trimmedCommand)); err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Write error for client: %s\n", err)
		clientConn.Close()
		return "", fmt.Errorf("write error for client: %s", err)
	}

	// Listen for a response
	_, response, err := clientConn.ReadMessage()
	if err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Read error for client: %s\n", err)
		return "", fmt.Errorf("read error for client %s: %v", ip_addr, err)
	}

	// Return the response as a string
	return string(response), nil
}

func postCommandAndListenInternal(clientConn *websocket.Conn, ip_addr string, command string) (string, error) {
	// Take in a command and ip address, and try across the client array for connecting by IP.
	trimmedCommand := strings.TrimSpace(command)
	clientTitle := client.ClientUser[ip_addr]

	// Post command
	// This will send the command across the web socket, and catch any errors that are sent back in response.
	// Theres no need to print the full response here, as the websocket will do that automatically in the main
	// loop.
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(trimmedCommand)); err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Write error for client: %s\n", err)
		clientConn.Close()
		return "", fmt.Errorf("write error for client: %s", err)
	}

	// Write our command to the output log
	ui.QueueMessageToBeWritten(ui.OutputLog, "%s %s\n", clientTitle, trimmedCommand)

	// Listen for a response
	_, response, err := clientConn.ReadMessage()
	if err != nil {
		ui.QueueMessageToBeWritten(ui.ErrorLog, "Read error for client: %s\n", err)
		return "", fmt.Errorf("read error for client %s: %v", ip_addr, err)
	}

	// Return the response as a string
	return string(response), nil
}

// Send a comand to all attached clients
func BulkRequest(command string, clientList map[string]*websocket.Conn) {
	fullCommand := strings.TrimSpace(command)
	ui.QueueMessageToBeWritten(ui.OutputLog, "[green]Command sent to clients [-][[blue]%s[-]]: %s\n", client.TargetClient, fullCommand)

	for ip_addr, clientConn := range clientList {
		if err := clientConn.WriteMessage(websocket.TextMessage, []byte(fullCommand)); err != nil {
			ui.QueueMessageToBeWritten(ui.ErrorLog, "Write error for client [%s]: %s", ip_addr, err)
			clientConn.Close()
		}
	}
}
