package client

import "github.com/gorilla/websocket"

// Map of all client websockets
var Clients = make(map[string]*websocket.Conn)
var ClientUser = make(map[string]string)

// String references
var TargetClient string
var WildcardPattern string
var WildcardArray map[string]*websocket.Conn
var AliasArray map[string]*websocket.Conn

// Add client
func RegisterClient(ip_addr string, title string) {
	ClientUser[ip_addr] = title
}

func AddClient(ip_addr string, socket *websocket.Conn) {
	Clients[ip_addr] = socket
}

func RemoveClientConnection(ip_addr string) {
	delete(Clients, ip_addr)
}

func RemoveClientUserConnect(ip_addr string) {
	delete(ClientUser, ip_addr)
}
