package state

import "testing"

func TestStateMachineImpls(t *testing.T) {
	csm := interface{}(&CommunityStateMachine{})
	_, ok := csm.(StateMachine)
	if !ok {
		t.Error("not a valid impl of StateMachine interface")
	}
}
