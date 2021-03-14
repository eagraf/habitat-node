package entities

// A ReplicatedStateMachine is an ordered log of deterministic operations that
// are agreed upon by a consensus algorithm between different nodes
type ReplicatedStateMachine interface {
	// Start() causes this node to join the cluster, and also perform associated operations
	// like restoring the log to the most updated state if operations were missed
	Start()

	// Propose() asks the cluster to accept a new operation in the next log slot.
	// Asking the RSM to apply an operation is not a guarantee that it will succeed,
	// since it will only be commited if consensus is reached
	Propose() error

	// Apply() is called when this node receives confirmation that an operation was commited
	// Monitoring processes are notified, passing on the Transition to TransitionSubscribers
	Apply() error
}

type ConsensusAlgorithmType string

const (
	Raft ConsensusAlgorithmType = "raft"
)

type ConsensusAlgorithmConfig struct {
	Type   ConsensusAlgorithmType `json:"type"`
	Config map[string]interface{} `json:"config"`
}
