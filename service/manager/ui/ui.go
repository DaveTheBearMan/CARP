package ui

import (
	"fmt"
	"http_proxy/manager/client"
	"http_proxy/manager/parser"
	"http_proxy/manager/utils"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

// tview Application and children
var (
	App                *tview.Application
	SelectedClient     *tview.TextView
	ActiveClients      *tview.TextView
	OutputLog          *tview.TextView
	ConnectionLog      *tview.TextView
	ErrorLog           *tview.TextView
	CommandField       *tview.InputField
	GlobalCommandField *tview.InputField
	InputField         *tview.InputField
	Flex               *tview.Flex
)

// Main function to handle when you enter a clients ip address / alias / wildcard
func handleEnterKeyPress() {
	ipAddr := InputField.GetText()
	InputField.SetText("") // Clear input field immediately

	switch {
	case ipAddr == "all":
		handleAllClientsSelection()
	case parser.CheckAddress(ipAddr)[0] != "":
		// Wipe the old alias array, and set our target client before handling selection
		// I am aware of the fact that this can cause memory problems. I honestly do not know well enough
		// how golands garbage collection works. I don't know if it automatically cleans this up. I don't know how this
		// will react performance wise as the C2 scales. But I do know that I wasn't willing to waste a bunch of time
		// on the most efficient way to clear an array.
		client.AliasArray = make(map[string]*websocket.Conn)
		client.TargetClient = "alias"
		handleAliasSelection(parser.CheckAddress(ipAddr))

		// Update selected client to be our alias, and clear output log.
		SelectedClient.SetText(fmt.Sprintf(" Alias: %s", ipAddr))
		OutputLog.Clear()

		// CLear active clients and rewrite with our matches
		ActiveClients.Clear()
		QueueMessageToBeWritten(ActiveClients, "%s", utils.WriteKeysFromMap(""))
	case strings.Contains(ipAddr, "*"):
		handleWildcardSelection(ipAddr)
	default:
		handleSpecificClientSelection(ipAddr)
	}
}

// Function to handle the basic client selection for all clients
func handleAllClientsSelection() {
	client.TargetClient = "all"
	SelectedClient.SetText(" All clients ")
	OutputLog.Clear()

	// Clear and update clients
	ActiveClients.Clear()
	QueueMessageToBeWritten(ActiveClients, "%s", utils.WriteKeysFromMap(""))
}

// Handles just wildcard selection for an ip address range
func handleWildcardSelection(pattern string) {
	matchingIps := utils.ParseForWildCard(client.Clients, pattern)
	client.TargetClient = "wildcard"
	client.WildcardArray = utils.CopyMap(matchingIps)
	client.WildcardPattern = pattern
	SelectedClient.SetText(" " + pattern)
	OutputLog.Clear()

	// Clear and update clients
	ActiveClients.Clear()
	QueueMessageToBeWritten(ActiveClients, "%s", utils.WriteKeysFromMap(""))
}

// Handles when an alias is selected recursively so that you get the entire stack
func handleAliasSelection(patterns []string) {
	for _, ipAddr := range patterns {
		// Check if we need to recurse, otherwise wildcard the ip and check for it.
		switch {
		case parser.CheckAddress(ipAddr)[0] != "":
			handleAliasSelection(parser.CheckAddress(ipAddr))
		default:
			for ip, conn := range utils.ParseForWildCard(client.Clients, ipAddr) {
				client.AliasArray[ip] = conn
			}
		}
	}
}

// Handles when we want exactly one client connected
func handleSpecificClientSelection(ipAddr string) {
	if _, exists := client.Clients[ipAddr]; exists {
		client.TargetClient = ipAddr
		SelectedClient.SetText(" " + ipAddr)
		OutputLog.Clear()

		// Clear and update clients
		ActiveClients.Clear()
		QueueMessageToBeWritten(ActiveClients, "%s", utils.WriteKeysFromMap(""))
	}
}

// Sets up the ui flex panel
func SetupUI() {
	// Selected Client
	SelectedClient = tview.NewTextView()
	SelectedClient.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Active Client ")

	// Active clients
	ActiveClients = tview.NewTextView()
	ActiveClients.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Clients ")

	// Connection log
	ConnectionLog = tview.NewTextView()
	ConnectionLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Connection Log ")

	// Error log
	ErrorLog = tview.NewTextView()
	ErrorLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Error Log ")

	// Input field for command name
	CommandField = tview.NewInputField()
	CommandField.SetTitle(" Send Command ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Input field for a global command
	GlobalCommandField = tview.NewInputField().
		SetAcceptanceFunc(tview.InputFieldMaxLength(25)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				GlobalCommandField.SetText(parser.ParseIncoming(GlobalCommandField.GetText()))
				// Spawn a thread to display our commands output.
				go func() {
					GlobalCommandField.SetDisabled(true)
					time.Sleep(2 * time.Second)
					GlobalCommandField.SetDisabled(false)
					GlobalCommandField.SetText("")
					App.Draw()
				}()
			}
		})
	GlobalCommandField.SetTitle(" Global Command ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Input field for client name
	InputField = tview.NewInputField().
		SetLabel(" Client IP ").
		SetLabelColor(tcell.ColorWhite).
		SetAcceptanceFunc(tview.InputFieldMaxLength(25)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				handleEnterKeyPress()
			}
		})

	InputField.SetTitle(" Select Client ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Text view to display the list of clients
	OutputLog = tview.NewTextView()
	OutputLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Response ")

	// Assemble the layout
	Flex = tview.NewFlex().
		AddItem(tview.NewFlex().
			AddItem(InputField, 0, 1, false).
			AddItem(SelectedClient, 0, 1, false).
			AddItem(ActiveClients, 0, 10, false).SetDirection(tview.FlexRow), 0, 2, true).
		AddItem(tview.NewFlex().
			AddItem(OutputLog, 0, 1, false).
			AddItem(CommandField, 3, 1, false).
			AddItem(GlobalCommandField, 3, 1, false).SetDirection(tview.FlexRow), 0, 5, false).
		AddItem(tview.NewFlex().
			AddItem(ConnectionLog, 0, 1, false).
			AddItem(ErrorLog, 0, 1, false).SetDirection(tview.FlexRow), 0, 2, false)
}
