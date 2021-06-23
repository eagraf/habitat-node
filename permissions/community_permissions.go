package permissions

import (
	"errors"

	"github.com/eagraf/habitat-node/entities"
)

// from https://www.davidkaya.com/sets-in-golang/
// should probably create a lib package for this type of stuff

var exists = struct{}{}

type set struct {
	m map[entities.UserID]struct{}
}

func NewSet() *set {
	s := &set{}
	s.m = make(map[entities.UserID]struct{})
	return s
}

func (s *set) Add(value entities.UserID) {
	s.m[value] = exists
}

func (s *set) Remove(value entities.UserID) {
	delete(s.m, value)
}

func (s *set) Contains(value entities.UserID) bool {
	_, c := s.m[value]
	return c
}

type CommunityPermissions struct {
	AdminCapabilities  Actions
	UserCapabilities   Actions
	ViewerCapabilities Actions
	Admins             *set
	Users              *set
	Viewers            *set
}

func (cp CommunityPermissions) GetAdminCapabilities() Actions {
	return cp.AdminCapabilities
}

func (cp CommunityPermissions) GetUserCapabilities() Actions {
	return cp.UserCapabilities
}

func (cp CommunityPermissions) GetViewerCapabilities() Actions {
	return cp.ViewerCapabilities
}

func BootstrapPermissions(user entities.User) *CommunityPermissions {
	s := NewSet()
	s.Add(user.ID)
	return &CommunityPermissions{
		AdminCapabilities:  make([]Action, 0),
		UserCapabilities:   make([]Action, 0),
		ViewerCapabilities: make([]Action, 0),
		Admins:             s,
		Users:              NewSet(),
		Viewers:            NewSet(),
	}
}

func getDefaultAdminCapabilities() Actions {
	return []Action{
		"ADD_USER", "ADD_VIEWER",
		"REMOVE_USER", "REMOVE_VIEWER",
		"ADD_APP", "REMOVE_APP",
		"MESSAGE_USER", "MESSAGE_VIEWER",
		"ADD_USER_CAPABILITY", "ADD_VIEWER_CAPABILITY",
		"REMOVE_USER_CAPABILITY", "REMOVE_VIEWER_CAPABILITY",
	}
}

func getDefaultUserCapabilities() Actions {
	return []Action{
		"INVITE_USER", "INVITE_VIEWER",
		"REQUEST_REMOVE_USER", "REQUEST_REMOVE_VIEWER",
		"ADD_APP", "REMOVE_APP",
		"MESSAGE_USER",
		"USE_APP",
	}
}

func getDefaultViewerCapabilities() Actions {
	return []Action{
		"USE_NONCOLL_APP",
	}
}

func (cp *CommunityPermissions) SetDefaultPermissions() {
	cp.AdminCapabilities = getDefaultAdminCapabilities()
	cp.UserCapabilities = getDefaultUserCapabilities()
	cp.ViewerCapabilities = getDefaultViewerCapabilities()
}

func (cp *CommunityPermissions) isValidAction(user entities.User, A Action) (bool, error) {
	if cp.Admins.Contains(user.ID) {
		return isValidAction(cp, A, Admin), nil
	} else if cp.Users.Contains(user.ID) {
		return isValidAction(cp, A, User), nil
	} else if cp.Viewers.Contains(user.ID) {
		return isValidAction(cp, A, Viewer), nil
	} else {
		return false, errors.New("User is niether in Admins nor Users nor Viewers for this group")
	}
}
