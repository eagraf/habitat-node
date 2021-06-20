package app

import (
	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
)

type CommunityManagerConfig struct {
	Community *entities.Community
	User      *entities.User
}

func (cm *CommunityManagerConfig) Receive(transitions.Transition) {
	return
}

func (cm *CommunityManagerConfig) ListMembers() (string, error) {
	return "", nil
}

func (cm *CommunityManagerConfig) AddMember() error {
	return nil
}

func (cm *CommunityManagerConfig) RemoveMember() error {
	return nil
}

func (cm *CommunityManagerConfig) BanMember() error {
	return nil
}

func (cm *CommunityManagerConfig) AssignPermissions(user *entities.User) {

}
