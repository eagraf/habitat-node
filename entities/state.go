package entities

// State holds critical state for a node.
// Different modules act on this struct as a state machine.
type State struct {
	Communities []Community `json:"communities"` // Communities that the node is helping to host
	Users       []User      `json:"users"`       // Users that are logged into this node
}
