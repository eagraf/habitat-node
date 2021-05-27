package transitions

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/eagraf/habitat-node/entities"
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
	HostCategory      TransitionSubscriptionCategory = "HOST"
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
	//Reduce(*entities.State) (*entities.State, error)
	// TODO we might need to implement rollbacks as well
	// TODO we might need to add a validate method
}

type CommunityTransition interface {
	Reduce(*entities.Community) (*entities.Community, error)
	//Rollback(*entities.Community) (*entities.Community, error)
	CommunityID() entities.CommunityID
}

type HostUserTransition interface {
	Reduce(*entities.HostUser) (*entities.HostUser, error)
}

type HostTransition interface {
	Reduce(*entities.Host) (*entities.Host, error)
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
