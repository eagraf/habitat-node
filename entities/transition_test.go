package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddCommunity(t *testing.T) {
	state := InitState()
	community := InitCommunity("community_0", "My Community", IPFS)

	transition := AddCommunityTransition{
		Community: community,
	}

	_, err := transition.Reduce(state)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(state.Communities))
}
