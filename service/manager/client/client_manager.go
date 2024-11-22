package client

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Map of all client websockets
var Clients = make(map[string]*websocket.Conn)
var ClientUser = make(map[string]string)

// String references
var TargetClient string
var WildcardPattern string
var WildcardArray map[string]*websocket.Conn
var AliasArray map[string]*websocket.Conn
var mu sync.Mutex // mutex for thread-safety

// Add client
func RegisterClient(ip_addr string, title string) {
	mu.Lock()
	defer mu.Unlock()
	ClientUser[ip_addr] = title
}

func AddClient(ip_addr string, socket *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	Clients[ip_addr] = socket
}

func RemoveClientConnection(ip_addr string) {
	mu.Lock()
	defer mu.Unlock()
	delete(Clients, ip_addr)
}

func RemoveClientUserConnect(ip_addr string) {
	mu.Lock()
	defer mu.Unlock()
	delete(ClientUser, ip_addr)
}
