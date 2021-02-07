package entities

import "fmt"

// TransitionType enumerates possible state transitions
type TransitionType string

// All possible TransitionTypes
const (
	AddCommunityTransitionType TransitionType = "ADD_COMMUNITY"
	AddMemberTransitionType    TransitionType = "ADD_MEMBER"
)

// A Transition transitions the state from one arrangement to another
// Each state transition is implemented via a reducer function
type Transition interface {
	Type() TransitionType
	Reduce(*State) (*State, error)
	// TODO we might need to implement rollbacks as well
}

// TransitionSubscriber receives state transitions from a state monitoring process
type TransitionSubscriber interface {
	Receive(transition Transition) error
}

// Host transitions are initiated by a user on the host node

type AddCommunityTransition struct {
	Community Community `json:"community"`
}

func (ac AddCommunityTransition) Type() TransitionType {
	return AddCommunityTransitionType
}

func (ac AddCommunityTransition) Reduce(state *State) (*State, error) {
	newState := *state
	if _, ok := state.Communities[ac.Community.ID]; ok {
		return nil, fmt.Errorf("community with id %s is already in state", ac.Community.ID)
	}
	newState.Communities[ac.Community.ID] = ac.Community
	return &newState, nil
}

// Within community transitions are agreed upon by consensus between community member nodes

type AddMemberTransition struct {
	Community CommunityID `json:"community_id"`
	User      User        `json:"user`
}

func (am AddMemberTransition) Type() TransitionType {
	return AddMemberTransitionType
}

func (am AddMemberTransition) Reduce(state *State) (*State, error) {
	newState := *state
	if _, ok := state.Communities[am.Community]; !ok {
		return nil, fmt.Errorf("no community with id %s in state", am.Community)
	}
	if _, ok := state.Communities[am.Community].Members[am.User.ID]; ok {
		return nil, fmt.Errorf("community %s already has member %s", am.Community, am.User)
	}
	newState.Communities[am.Community].Members[am.User.ID] = am.User
	return &newState, nil
}
