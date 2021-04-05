package state

import (
	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
)

type StateMachine interface {
	Apply(transition transitions.CommunityTransition) error
}

type CommunityStateMachine struct {
	CommunityID   entities.CommunityID
	WriteAheadLog *Log
	State         *entities.Community
}

type HostStateMachine struct {
}

func (sm *CommunityStateMachine) Apply(transition transitions.CommunityTransition) error {
	return nil
}
