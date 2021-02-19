package entities

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// TransitionType enumerates possible state transitions
type TransitionType string

// All possible TransitionTypes
const (
	AddCommunityTransitionType  TransitionType = "ADD_COMMUNITY"
	AddMemberTransitionType     TransitionType = "ADD_MEMBER"
	UpdateBacknetTransitionType TransitionType = "UPDATE_BACKNET"
)

var transitionReflectionTypeRegistry = map[TransitionType]reflect.Type{
	AddCommunityTransitionType:  reflect.TypeOf(AddCommunityTransition{}),
	AddMemberTransitionType:     reflect.TypeOf(AddMemberTransition{}),
	UpdateBacknetTransitionType: reflect.TypeOf(UpdateBacknetTransition{}),
}

// TransitionSubscriptionCategory enumerates different types entities that a TransitionSubscriber could be subscribed to
type TransitionSubscriptionCategory string

// All possible TransitionSubscriptionCategories
const (
	CommunityCategory TransitionSubscriptionCategory = "COMMUNITY"
	HostUserCategory  TransitionSubscriptionCategory = "HOST_USER"
)

var transitionSubscriptionCategories = map[TransitionType]TransitionSubscriptionCategory{
	AddCommunityTransitionType:  CommunityCategory,
	AddMemberTransitionType:     CommunityCategory,
	UpdateBacknetTransitionType: CommunityCategory,
}

// A Transition transitions the state from one arrangement to another
// Each state transition is implemented via a reducer function
type Transition interface {
	Type() TransitionType
	Reduce(*State) (*State, error)
	// TODO we might need to implement rollbacks as well
	// TODO we might need to add a validate method
}

// TransitionWrapper adds type information to a transition in marshalled form
type TransitionWrapper struct {
	Type       TransitionType `json:"type"`
	Transition Transition     `json:"transition"`
}

// UnmarhsalJSON uses reflection to extract the proper Transition type from a TransitionWrapper JSON
func (tw *TransitionWrapper) UnmarshalJSON(bytes []byte) error {
	// first read into map[string]interface{} to extract type information
	var firstPass map[string]interface{}
	err := json.Unmarshal(bytes, &firstPass)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	transitionType, ok := firstPass["type"].(string)
	if !ok {
		return errors.New("Transitionwrapper json has no field type")
	}

	if _, ok := firstPass["transition"]; !ok {
		return errors.New("TransitionWrapper json has no field transition")
	}

	// Get the proper type to assert from the Transition type registry
	reflectType, ok := transitionReflectionTypeRegistry[TransitionType(transitionType)]
	if !ok {
		return fmt.Errorf("type %s not in transitionReflectionTypeRegistry", transitionType)
	}

	// Create a new struct pointer that is the correct implementation of Transition
	// and then decode into it using mapstructure
	transitionValue := reflect.New(reflectType)
	transition := transitionValue.Interface()
	err = mapstructure.Decode(firstPass["transition"], transition)
	if err != nil {
		return err
	}

	// Write into the receiving structs fields
	tw.Type = TransitionType(transitionType)
	tw.Transition, ok = transition.(Transition)
	if !ok {
		return errors.New("unmarshalled struct is not a Transition interface")
	}

	return nil
}

// CommunityTransition is a transition to a community element in state
type CommunityTransition interface {
	Transition
	CommunityID() CommunityID
}

// HostUserTransition is a transition to a host user element in state
type HostUserTransition interface {
	Transition
	HostUsername() string
}

// TransitionSubscriber receives state transitions from a state monitoring process
type TransitionSubscriber interface {
	Receive(transition Transition) error
}

// GetSubscriptionCategory returns the subscription category for a given transition type
func GetSubscriptionCategory(transitionType TransitionType) (TransitionSubscriptionCategory, error) {
	category, ok := transitionSubscriptionCategories[transitionType]
	if !ok {
		return "", fmt.Errorf("transition type %s not supported", string(transitionType))
	}
	return category, nil
}

// Host transitions are initiated by a user on the host node

type AddCommunityTransition struct {
	Community *Community `json:"community"`
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

func (ac AddCommunityTransition) CommunityID() CommunityID {
	return ac.Community.ID
}

// Within community transitions are agreed upon by consensus between community member nodes

type AddMemberTransition struct {
	Community CommunityID `json:"community_id"`
	User      *User       `json:"user`
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

func (am AddMemberTransition) CommunityID() CommunityID {
	return am.Community
}

type UpdateBacknetTransition struct {
	OldCommunity *Community
	NewCommunity *Community
}

func (ub UpdateBacknetTransition) Type() TransitionType {
	return UpdateBacknetTransitionType
}

func (ub UpdateBacknetTransition) Reduce(state *State) (*State, error) {
	newState := *state
	if _, ok := state.Communities[ub.OldCommunity.ID]; !ok {
		return nil, fmt.Errorf("no community with id %s in state", ub.OldCommunity.ID)
	}
	if ub.OldCommunity.ID != ub.NewCommunity.ID {
		return nil, fmt.Errorf("old and new community ids do not match %s, %s", ub.OldCommunity.ID, ub.NewCommunity.ID)
	}

	if ub.OldCommunity.Backnet.Type != ub.NewCommunity.Backnet.Type {
		return nil, fmt.Errorf("switching backnet implementations is not supported")
	}

	*newState.Communities[ub.OldCommunity.ID].Backnet = *ub.NewCommunity.Backnet

	return &newState, nil
}
