package ui

import (
	"http_proxy/manager/client"
	"http_proxy/manager/utils"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// tview Application and children
var (
	App            *tview.Application
	SelectedClient *tview.TextView
	ActiveClients  *tview.TextView
	OutputLog      *tview.TextView
	ConnectionLog  *tview.TextView
	ErrorLog       *tview.TextView
	CommandField   *tview.InputField
	Flex           *tview.Flex
)

func RedrawClientList() {
	if client.TargetClient == "wildcard" {

	}
}

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

	// Input field for client name
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel(" Client IP ").
		SetLabelColor(tcell.ColorWhite).
		SetAcceptanceFunc(tview.InputFieldMaxLength(20)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				ip_addr := inputField.GetText()
				if ip_addr == "all" {
					client.TargetClient = "all"
					inputField.SetText("")
					OutputLog.Clear()
					SelectedClient.SetText(" All clients ")
					return
				}

				if strings.Contains(ip_addr, "*") {
					matchingIps := utils.ParseForWildCard(client.Clients, ip_addr)
					client.TargetClient = "wildcard"
					client.WildcardArray = utils.CopyMap(matchingIps)
					client.WildcardPattern = ip_addr
					SelectedClient.SetText(" " + ip_addr)
					OutputLog.Clear()
					inputField.SetText("")
					ActiveClients.Clear()
					QueueMessageToBeWritten(ActiveClients, "%s", utils.WriteKeysFromMap(""))
				} else {
					_, ok := client.Clients[ip_addr]
					if ok {
						client.TargetClient = ip_addr
						SelectedClient.SetText(" " + ip_addr)
						inputField.SetText("")
						OutputLog.Clear()
					} else {
						inputField.SetText("")
					}
				}
			}
		})

	inputField.SetTitle(" Select Client ").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Text view to display the list of clients
	OutputLog = tview.NewTextView()
	OutputLog.SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Response ")

	// Assemble the layout
	Flex = tview.NewFlex().
		AddItem(tview.NewFlex().
			AddItem(inputField, 0, 1, false).
			AddItem(SelectedClient, 0, 1, false).
			AddItem(ActiveClients, 0, 10, false).SetDirection(tview.FlexRow), 0, 2, true).
		AddItem(tview.NewFlex().
			AddItem(OutputLog, 0, 10, false).
			AddItem(CommandField, 0, 1, false).SetDirection(tview.FlexRow), 0, 5, false).
		AddItem(tview.NewFlex().
			AddItem(ConnectionLog, 0, 1, false).
			AddItem(ErrorLog, 0, 1, false).SetDirection(tview.FlexRow), 0, 2, false)
}
