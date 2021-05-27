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
	}
}
	`)

	var tw TransitionWrapper
	err := json.Unmarshal(jsonBytes, &tw)
	assert.Nil(t, err)

	act, ok := tw.Transition.(*AddCommunityTransition)
	assert.Equal(t, true, ok)
	assert.Equal(t, "community_0", string(act.Community.ID))
}
