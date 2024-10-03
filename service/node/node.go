package main

// Utilizing gorilla mux for routers for the REST Api
import (
	"encoding/json"
	"http_proxy/types"
	"http_proxy/utils"
	"net/http"

	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Global Manager
var node = types.Node{
	Base: types.Base{
		UUID: uuid.New(),
		Address: types.Address{
			Ipv4: utils.GetOutboundIP().String(),
			Port: 8080,
		},
	},
	Clients: make(map[string]types.Client),
}

// Home page for the REST API
func landingPage(writer http.ResponseWriter, request *http.Request) {
	// Guarantee header and write current manager
	writer.Header().Set("Content-Type", "application/json")

	// Encode and write, catch the error
	err := json.NewEncoder(writer).Encode(node)
	if err != nil {
		http.Error(writer, "Error encoding JSON", http.StatusInternalServerError)
	}

	// Print for success or error
	utils.WrapErrorCheck(request, err, "Accessed landing page")
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

	// Split address into IP and port.
	addressPortCombo := strings.Split(request.RemoteAddr, ":")
	port, portErr := strconv.Atoi(addressPortCombo[1])

	// Check and see if we correctly recieved a port to communicate across
	if portErr != nil {
		utils.WrapErrorCheck(request, err, "Unable to translate port to integer")
		http.Error(writer, "Unable to translate port from address to integer", http.StatusInternalServerError)
		return
	}

	// Set the address of the client
	client.Address = types.Address{
		Ipv4: addressPortCombo[0],
		Port: port,
	}

	// Update nodes client table
	node.Clients[client.Address.Ipv4] = client

	// Write back the newly registered node
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(client)
}

func main() {
	router := mux.NewRouter()

	// Api access routes
	router.HandleFunc("/", landingPage).Methods("GET")
	router.HandleFunc("/register-client", registerClient).Methods("POST")

	// Open server
	localIP := utils.GetOutboundIP().String()
	utils.LogMessage(localIP, "Beginning listening on port 8080 . . .")
	http.ListenAndServe(":8080", router)
}
