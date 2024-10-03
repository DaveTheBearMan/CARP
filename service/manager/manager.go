package main

// Utilizing gorilla mux for routers for the REST Api
import (
	"encoding/json"
	"http_proxy/types"
	"http_proxy/utils"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Global Manager
var manager = types.Manager{
	Nodes:   make(map[uuid.UUID]types.Node),
	Clients: make(map[string]types.Client),
}

// Home page for the REST API
func landingPage(writer http.ResponseWriter, request *http.Request) {
	// Guarantee header and write current manager
	writer.Header().Set("Content-Type", "application/json")

	// Encode and write, catch the error
	err := json.NewEncoder(writer).Encode(manager)
	if err != nil {
		http.Error(writer, "Error encoding JSON", http.StatusInternalServerError)
	}

	// Print for success or error
	utils.WrapErrorCheck(request, err, "Accessed landing page")
}

// Register node
func registerNode(writer http.ResponseWriter, request *http.Request) {
	// Create a new node instance
	var node types.Node

	// Catch and return any errors over HTTP
	err := json.NewDecoder(request.Body).Decode(&node)
	utils.WrapErrorCheck(request, err, "Successfully registered new node")
	if err != nil {
		http.Error(writer, "Invalid input data", http.StatusBadRequest)
		return
	}

	// Update the manager
	node.UUID = uuid.New()
	manager.Nodes[node.UUID] = node

	// Write back the newly registered node
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(node)
}

// Register client
func registerClient(writer http.ResponseWriter, request *http.Request) {
	// Create a new node instance
	var client types.Client

	// Catch and return any errors over HTTP
	err := json.NewDecoder(request.Body).Decode(&client)
	utils.WrapErrorCheck(request, err, "Successfully registered new client")
	if err != nil {
		http.Error(writer, "Invalid input data", http.StatusBadRequest)
		return
	}

	// Update the manager
	manager.Clients[client.Address.Ipv4] = client

	// Write back the newly registered node
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(client)
}

func main() {
	router := mux.NewRouter()

	// Api access routes
	router.HandleFunc("/", landingPage).Methods("GET")
	router.HandleFunc("/register-node", registerNode).Methods("POST")
	router.HandleFunc("/register-client", registerClient).Methods("POST")

	// Open server
	localIP := utils.GetOutboundIP().String()
	utils.LogMessage(localIP, "Beginning listening on port 8080 . . .")
	http.ListenAndServe(":8080", router)
}
