package transitions

import (
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestModifyCommunityMembers(t *testing.T) {
	community := entities.InitCommunity("community_0", "My Community", entities.IPFS)
	user := entities.InitUser("uniqueid", "userhandle")

	transition := ModifyCommMembersTransition{
		Community: community,
		User:      user,
		ModType:   AddMember,
	}

	newComm, err := transition.Reduce(community)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(newComm.Members))
	val, ok := newComm.Members[user.ID]
	assert.NotNil(t, val)
	assert.True(t, ok)
	assert.Equal(t, user, val)

	transition = ModifyCommMembersTransition{
		Community: newComm,
		User:      user,
		ModType:   RemoveMember,
	}

	newComm, err = transition.Reduce(newComm)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(newComm.Members))

	transition = ModifyCommMembersTransition{
		Community: newComm,
		User:      user,
		ModType:   BanMember,
	}

	newComm, err = transition.Reduce(newComm)
	assert.NotNil(t, err)

}
