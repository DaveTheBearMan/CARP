package ui

import (
	"fmt"
	"log"

	"github.com/rivo/tview"
)

// logEntry struct to hold the destination and message
type logEntry struct {
	location *tview.TextView // The destination TextView
	message  string          // The message to log
}

// Buffered channel for logs
var logChan = make(chan logEntry, 100)

// Set up logging goroutine.
func init() {
	go func() {
		// Loop over each entry in the log channel, and attempt to write to that buffer.
		for entry := range logChan {
			if _, err := fmt.Fprint(entry.location, entry.message); err != nil {
				log.Printf("Error writing to location: %v", err)
				QueueMessageToBeWritten(ErrorLog, "Error writing to location: %v", err)
			}
			entry.location.ScrollToEnd()
			// Redraw the app to ensure updates come across.
			App.Draw()
		}
	}()
}

// Take in a place to update and a string/format, then run app.Draw on it
func QueueMessageToBeWritten(location *tview.TextView, baseString string, args ...any) {
	message := fmt.Sprintf(baseString, args...)
	logChan <- logEntry{location: location, message: message}
}

// // Write out our current user, directory, and host
// func GetClientInformation(clientIP string) {
// 	// Check if the index exists, and if not, create it. Avoid rerunning commands for hosts
// 	caller, ok := client.ClientUser[clientIP]
// 	if !ok {
// 		// Get user, host
// 		user, err := cmd.ReadCommandFromSocket(client.Clients[clientIP], "%s", "whoami")
// 		if err != nil {
// 			QueueMessageToBeWritten(OutputLog, "%s", "Unable to get whoami")
// 			return
// 		}
// 		user = strings.ReplaceAll(user, "\n", "")
// 		user = strings.ReplaceAll(user, "\t", "")
// 		user = strings.ReplaceAll(user, " ", "")
// 		hostname, err := cmd.ReadCommandFromSocket(clientIP, "hostname")
// 		if err != nil {
// 			QueueMessageToBeWritten(OutputLog, "%s", "Unable to get hostname")
// 			return
// 		}
// 		hostname = strings.ReplaceAll(hostname, "\n", "")
// 		hostname = strings.ReplaceAll(hostname, "\t", "")
// 		hostname = strings.ReplaceAll(hostname, " ", "")

// 		client.ClientUser[clientIP] = fmt.Sprintf("[green]%s@%s[-]: $ ", user, hostname)
// 	} else {
// 		// Write to the output buffer the caller
// 		QueueMessageToBeWritten(OutputLog, caller)
// 	}
// }
