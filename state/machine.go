package state

import "github.com/eagraf/habitat-node/entities"

type StateMachine interface {
	Apply(transition entities.CommunityTransition) error
}

type CommunityStateMachine struct {
	CommunityID   entities.CommunityID
	WriteAheadLog *Log
	State         *entities.Community
}

type HostStateMachine struct {
}

func (sm *CommunityStateMachine) Apply(transition entities.CommunityTransition) error {
	return nil
}
