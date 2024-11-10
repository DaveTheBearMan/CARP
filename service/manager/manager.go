package main

import (
	"http_proxy/manager/cmd"
	"http_proxy/manager/ui"
	"log"
	"net/http"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Main function
func main() {
	// Log that the application is starting
	log.Println("Starting application...")

	// Start the HTTP server for WebSocket connections
	http.HandleFunc("/ws", cmd.HandleWebSocket)
	ui.App = tview.NewApplication()
	ui.SetupUI()
	log.Println("UI setup complete.")

	// Start the HTTP server in a goroutine
	go func() {
		log.Println("Starting HTTP server on :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	// Configure the CommandField to run commands
	if ui.CommandField != nil {
		ui.CommandField.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				cmd.RunCommand(ui.CommandField)
				log.Println("Command executed:", ui.CommandField.GetText())
			}
		})
	} else {
		log.Fatal("CommandField is not initialized")
	}

	// Start the TUI application
	log.Println("Starting TUI application...")
	if err := ui.App.SetRoot(ui.Flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}