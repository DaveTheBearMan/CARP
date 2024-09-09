package main

// Utilizing gorilla mux for routers for the REST Api
import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"http_proxy/utils"
)

// Handle client nodes
type Node struct {
	UUID	uuid.UUID		`json:"uuid"`
	Addr	string			`json:"address"`
	Clients	map[string]string	`json:"clients"`
}

// Node Holder
type Management struct {
	UUID	uuid.UUID		`json:"uuid"`
	Nodes	map[string]Node		`json:"nodes"`
}

// Global Manager
var globalManager = Management{
	UUID: uuid.New(),
	Nodes: make(map[string]Node),
}

// Home page for the REST API
func landingPage(writer http.ResponseWriter, request *http.Request) {
	// Guarantee header and write current manager
	writer.Header().Set("Content-Type", "application/json")

	// Encode and write, catch the error
	err := json.NewEncoder(writer).Encode(globalManager)
	if err != nil {
		http.Error(writer, "Error encoding JSON", http.StatusInternalServerError)
	}

	// Print for success or error
	utils.WrapErrorCheck(request, err, "Accessed landing page")
}

// Register node
func registerNode(writer http.ResponseWriter, request *http.Request) {
	// Create a new node instance
	var newNode Node

	// Catch and return any errors over HTTP
	err := json.NewDecoder(request.Body).Decode(&newNode)
	utils.WrapErrorCheck(request, err, "Successfully registered new node")
	if err != nil {
		http.Error(writer, "Invalid input data", http.StatusBadRequest)
		return
	}

	// Update the node
	newNode.UUID = uuid.New()
	newNode.Addr = utils.GetIPFromRequest(request)
	newNode.Clients = make(map[string]string)
	globalManager.Nodes[newNode.Addr] = newNode

	// Write back the newly registered node
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(newNode)
}

func main() {
	router := mux.NewRouter()

	// Api access routes
	router.HandleFunc("/", landingPage).Methods("GET")
	router.HandleFunc("/register-node", registerNode).Methods("POST")

	// Open server
	localIP := utils.GetOutboundIP().String()
	utils.LogMessage(localIP, "Beginning listening on port 8080 . . .")
	http.ListenAndServe(":8080", router)
}
