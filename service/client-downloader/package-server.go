package main

import (
	"log"
	"net/http"
)

func main() {
	// Set the file path to your binary
	filePath := "../node/node"

	// Handle the download request
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename=manager")
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, filePath)
	})

	// Start the server
	log.Println("Starting server on :5542")
	err := http.ListenAndServe(":5542", nil)
	if err != nil {
		log.Fatal(err)
	}
}
