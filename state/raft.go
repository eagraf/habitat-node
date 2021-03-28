package state

import (
	"io"

	"github.com/hashicorp/raft"
)

// RaftRSM implements the ReplicatedStateMachine interface
type RaftRSM struct {
}

func (r *RaftRSM) Start() {
	// Check stable storage for log/snapshots
}

func (r *RaftRSM) Propose() error {
	return nil
}

func (r *RaftRSM) Apply() error {
	return nil
}

type fsm RaftRSM

func (f *fsm) Apply(*raft.Log) interface{} {
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (f *fsm) Restore(io.ReadCloser) error {
	return nil
}
