package permissions

type RoleType string

const (
	Admin  RoleType = "ADMIN"
	User   RoleType = "USER"
	Viewer RoleType = "VIEWER"
)

type Action string
type Actions []Action

type Permissions interface {
	GetAdminCapabilities() Actions
	GetUserCapabilities() Actions
	GetViewerCapabilities() Actions
}

func (s Actions) contains(str Action) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func isValidAction(P Permissions, A Action, R RoleType) bool {
	switch R {
	case Admin:
		if P.GetAdminCapabilities().contains(A) {
			return true
		}
	case User:
		if P.GetUserCapabilities().contains(A) {
			return true
		}
	case Viewer:
		if P.GetViewerCapabilities().contains(A) {
			return true
		}
	}
	return false
}
