package transitions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalTransitionWrapper(t *testing.T) {
	jsonBytes := []byte(`
{
	"type": "ADD_COMMUNITY",
	"transition": {
		"community": {
			"id": "community_0",
			"name": "my community"
		}
	},
	"sequence_number": 1
}
	`)

	var tw TransitionWrapper
	err := json.Unmarshal(jsonBytes, &tw)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), tw.SequenceNumber)

	act, ok := tw.Transition.(*AddCommunityTransition)
	assert.Equal(t, true, ok)
	assert.Equal(t, "community_0", string(act.Community.ID))
}
