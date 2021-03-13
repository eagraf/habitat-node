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
}

func (r *RaftRSM) Apply() error {
}

type fsm RaftRSM

func (f *fsm) Apply(*raft.Log) interface{} {

}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {

}

func (f *fsm) Restore(io.ReadCloser) error {

}
