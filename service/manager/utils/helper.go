package utils

import (
	"bytes"
	"fmt"
	"http_proxy/manager/client"
	"log"
	"net"
	"strings"

	"github.com/gorilla/websocket"
)

// Takes in a map containing IP address indices and websocket connections, and returns all available IPs.
func WriteKeysFromMap(prefix string) string {
	b := new(bytes.Buffer)
	written := make(map[string]bool)

	// Make matching wildcards red
	if client.TargetClient == "wildcard" {
		for key := range client.WildcardArray {
			written[key] = true
			fmt.Fprintf(b, "%s[red]%s[-]\n", prefix, key)
		}
	}
	for key := range client.Clients {
		if !written[key] {
			fmt.Fprintf(b, "%s%s\n", prefix, key)
		}
	}

	return b.String()
}

// Take in a map of IP addresses, and a wildcard address, and accept any that match.
func ParseForWildCard(m map[string]*websocket.Conn, wildcard string) map[string]*websocket.Conn {
	matchingAddresses := make(map[string]*websocket.Conn)
	splitWildcardAddress := strings.Split(wildcard, ".")

	// Iterate over passed in IP addresses and attempt to match.
	for ip, conn := range m {
		matching := true
		splitHostAddress := strings.Split(ip, ".")

		for i := 0; i <= 3; i++ {
			if splitWildcardAddress[i] != "*" && splitWildcardAddress[i] != splitHostAddress[i] {
				matching = false
				break
			}
		}

		if matching {
			matchingAddresses[ip] = conn
		}
	}

	return matchingAddresses
}

// Get outbound IP address
func GetOutboundIPAddress() string {
	Connection, Err := net.Dial("udp", "8.8.8.8:80")
	if Err != nil {
		log.Fatal(Err)
	}
	defer Connection.Close()
	localAddress := Connection.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.String()
}

// Copy web socket map.
func CopyMap(original map[string]*websocket.Conn) map[string]*websocket.Conn {
	copied := make(map[string]*websocket.Conn)
	for key, value := range original {
		copied[key] = value
	}
	return copied
}
