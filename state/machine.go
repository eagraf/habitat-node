package state

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
)

type StateMachine interface {
	Restart() error
	Apply(transition transitions.CommunityTransition) error
	// TODO snapshot
}

type CommunityStateMachine struct {
	CommunityID   entities.CommunityID
	WriteAheadLog *Log
	State         *entities.Community
	Path          string
}

type HostStateMachine struct {
}

// InitCommunityStateMachine gets ready for a restart or for a clean start
func InitCommunityStateMachine(communityID entities.CommunityID, stateBaseDir string) (*CommunityStateMachine, error) {
	stateDir := filepath.Join(stateBaseDir, string(communityID))

	// Check if state machine dir exists
	_, err := os.Stat(stateDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Start new state machine dir
	if os.IsNotExist(err) {
		err := os.MkdirAll(stateDir, 0744)
		if err != nil {
			return nil, err
		}
	}

	// Initialize log
	log, err := NewLog(filepath.Join(stateDir, "wal"))
	if err != nil {
		return nil, err
	}

	// The state variable is not reconstituted just yet
	return &CommunityStateMachine{
		CommunityID:   communityID,
		WriteAheadLog: log,
		Path:          stateDir,
	}, nil
}

func (sm *CommunityStateMachine) Restart() error {
	// Reconstitute state from snapshot
	snapshotFile, err := os.OpenFile(filepath.Join(sm.Path, "snapshot"), os.O_RDONLY, 0644)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	defer snapshotFile.Close()

	var snapshotState entities.Community
	if !os.IsNotExist(err) {
		_, err := ReadSnapshot(snapshotFile, &snapshotState)
		if err != nil {
			return err
		}
	}

	// Roll up logs with sequence number higher than snapshot
	entries, err := sm.WriteAheadLog.GetEntries()
	if err != nil {
		return err
	}

	intermediateState := &snapshotState
	for _, entry := range entries {
		transition, ok := entry.Transition.Transition.(transitions.CommunityTransition)
		if !ok {
			return errors.New("transition in log entry was not a CommunityTransition")
		}
		tempState, err := transition.Reduce(intermediateState)
		if err != nil {
			return err
		}
		intermediateState = tempState
	}

	sm.State = intermediateState

	// TODO restart consensus algorithm

	return nil
}

func (sm *CommunityStateMachine) Apply(transition transitions.CommunityTransition) error {
	// Apply reducer to state
	newState, err := transition.Reduce(sm.State)
	if err != nil {
		return err
	}

	sm.State = newState

	// TODO snapshot if needed

	// TODO notify all transition subscribers

	return nil
}
