package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
	"github.com/rs/zerolog/log"
)

type StateMachine interface {
	Restart() error
	Apply(transition *transitions.TransitionWrapper) error
	GetState() interface{}
	// TODO snapshot
}

type CommunityStateMachine struct {
	CommunityID       entities.CommunityID
	WriteAheadLog     *Log
	CurSequenceNumber uint64
	State             *entities.Community
	Path              string
	SnapshotInterval  int
}

type HostStateMachine struct {
}

// InitCommunityStateMachine gets ready for a restart or for a clean start
func InitCommunityStateMachine(communityID entities.CommunityID, stateBaseDir string, snapshotInterval int) (*CommunityStateMachine, error) {
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
		CommunityID:      communityID,
		WriteAheadLog:    log,
		Path:             stateDir,
		SnapshotInterval: snapshotInterval,
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
	var snapshot *Snapshot
	if !os.IsNotExist(err) {
		s, err := ReadSnapshot(snapshotFile, &snapshotState)
		if err != nil {
			return err
		}
		snapshot = s
	}

	// Roll up logs with sequence number higher than snapshot
	entries, err := sm.WriteAheadLog.GetEntries()
	if err != nil {
		return err
	}

	// Validate that sequence numbers match
	expectedSequenceNumber := snapshot.SequenceNumber + uint64(len(entries))
	actualSequenceNumber := entries[len(entries)-1].SequenceNumber
	if expectedSequenceNumber != actualSequenceNumber {
		return fmt.Errorf("sequence number mismatch: %d expected, got %d", expectedSequenceNumber, actualSequenceNumber)
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
	sm.CurSequenceNumber = actualSequenceNumber

	// TODO restart consensus algorithm

	return nil
}

func (sm *CommunityStateMachine) Apply(transition *transitions.TransitionWrapper) error {
	// Validate that the transition is a community transition
	category, err := transitions.GetSubscriptionCategory(transition.Type)
	if err != nil {
		return err
	} else if category != transitions.CommunityCategory {
		return fmt.Errorf("transition category is %s, should be %s", category, transitions.CommunityCategory)
	}

	// Validate sequence number for new transition
	if transition.SequenceNumber != sm.CurSequenceNumber+1 {
		return fmt.Errorf("sequence number %d is off, should be %d", transition.SequenceNumber, sm.CurSequenceNumber+1)
	}

	// Write to write ahead log
	err = sm.WriteAheadLog.WriteAhead(transition)
	if err != nil {
		return err
	}

	// Very important that this is incremented immediately after write to write ahead log succeeds
	sm.CurSequenceNumber += 1

	// Apply reducer to state
	newState, err := transition.Transition.(transitions.CommunityTransition).Reduce(sm.State)
	if err != nil {
		return err
	}

	sm.State = newState

	// Copy snapshot file
	if sm.CurSequenceNumber%uint64(sm.SnapshotInterval) == 0 {
		err := ArchiveSnapshotFile(sm.Path, sm.SnapshotInterval)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		snapshotPath := filepath.Join(sm.Path, "snapshot")
		snapshotFile, err := os.OpenFile(snapshotPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		err = WriteSnapshot(snapshotFile, sm.State, sm.CurSequenceNumber)
		if err != nil {
			return err
		}
	}

	// TODO notify all transition subscribers

	return nil
}

func (sm *CommunityStateMachine) GetState() interface{} {
	return sm.State
}
