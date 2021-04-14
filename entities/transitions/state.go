package transitions

import "github.com/eagraf/habitat-node/entities"

// state is a state object that is materialized by a sequence of transitions

type State interface {
	Reduce(Transition) error
	Rollback(Transition) error
}

type CommunityState struct {
	Community entities.Community `json:"community"`
}

type HostUserState struct {
}
