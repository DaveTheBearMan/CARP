package types

import (
	"github.com/google/uuid"
)

type Address struct {
	Ipv4 string `json:"ipv4"`
	Port int    `json:"port"`
}

type Base struct {
	UUID    uuid.UUID `json:"uuid"`
	Address Address   `json:"address"`
}

type Client struct {
	Base           // Inherit base class
	Node uuid.UUID `json:"node"`
}

type Node struct {
	Base                      // Inherit base class
	Clients map[string]Client `json:"clients"`
}

type Manager struct {
	Nodes   map[uuid.UUID]Node `json:"nodes"`
	Clients map[string]Client  `json:"clients"`
}
