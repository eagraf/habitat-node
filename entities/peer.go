package entities

// Peer represents a node in the p2p network.
type Peer struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}
