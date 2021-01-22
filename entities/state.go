package entities

// State holds critical state for a node.
// Different modules act on this struct as a state machine.
type State struct {
	Communities map[CommunityID]Community `json:"communities"` // Communities that the node is helping to host
	Users       map[UserID]User           `json:"users"`       // Users that are logged into this node
}

func InitState() *State {
	return &State{
		Communities: make(map[CommunityID]Community),
		Users:       make(map[UserID]User),
	}
}
