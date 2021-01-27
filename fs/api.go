package fs

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
)

// Session is a logged in user
// TODO: move to somewhere else?
type Session struct {
	CommunityID entities.CommunityID
	User        *client.User
}

// Permission is either
// 1. a string list of all users with access (Acess Control List) or
// 2. a bool where its existence indicates that all users have access (value doesn't matter)
type Permission interface {
	isPerm()
}

// ACL is an access control list
type ACL []string        // for now its a string list of user ids, should probably be a bit mask for minimal storage
func (acl *ACL) isPerm() {}

// All is an indicator that everyone in the community has access
type All bool         // maybe should make this const/strings?
func (a All) isPerm() {}

// FilePermission stores permissions for a file
type FilePermission struct {
	readPermissions  Permission
	writePermissions Permission
}

// FileMetadata stores file metadata
// don't really know what else goes in here for now
type FileMetadata struct {
	permissions FilePermission
	lastEdit    string
}

// FilesystemService just needs authService for now
type FilesystemService struct {
	authService *client.AuthService
	state       *entities.State
	// i want this to be the receiver, not auth service, although that might be all we need (for now)
	nets map[entities.CommunityID]Backnet
}

// NewFilesystemService initializes the FS service given an auth service
func NewFilesystemService(as *client.AuthService, s *entities.State, n map[entities.CommunityID]Backnet) (*FilesystemService, error) {

	res := &FilesystemService{
		authService: as,
		state:       s,
		nets:        n,
	}
	return res, nil

}

// TODO: actually do something here!

// AuthenticateSessionToken takes in a token and returns the user
// maybe return a session?
func (fs *FilesystemService) AuthenticateSessionToken(token string) (*client.User, error) {
	// check token db to see if it exists - if so return user
	user, err := fs.authService.CheckToken(token)
	if err != nil {
		return nil, err
	}
	return user, nil

}

// CheckPermissions checks permissions given a user, community and path
// TODO: fix; just returns true for now
func CheckPermissions(userid *client.User, groupID entities.CommunityID, path string) (bool, error) {
	return true, nil
}

// GetRequestQueries takes in a http request and returns arguments provided in it
func GetRequestQueries(r *http.Request) url.Values {
	args := r.URL.Query()
	return args
}

// ParseFilePath path takes in a "url type path" and
// if it is in the format <community_id>:<file_path> returns community_id, file_path
// else it returns error
func ParseFilePath(path string) (entities.CommunityID, string, error) {
	re := regexp.MustCompile(`^([0-9]*):(/[^/ ]*)+/?$`)
	if (re.MatchString(path)) == false {
		return "", "", errors.New("invalid path format")
	}

	col := regexp.MustCompile(`:`)
	arr := col.Split(path, 1)

	if len(arr) < 2 {
		return "", "", errors.New("invalid path format")
	}
	return entities.CommunityID(arr[0]), arr[1], nil // is casting bad?

}

// DoChecks checks the session token and the permissions and either returns a session and filepath or error
func (fs *FilesystemService) DoChecks(args url.Values) (*Session, string, error) {
	token := args.Get("token")
	if token == "" {
		return nil, "", errors.New("could not find the session token argument in the URL")
	}

	user, err := fs.AuthenticateSessionToken(token)
	if err != nil {
		return nil, "", err
	}

	path := args.Get("path")
	if path == "" {
		return nil, "", errors.New("This path does not exist in the given groupspace")
	}

	commID, filepath, err := ParseFilePath(path)

	allow, err := CheckPermissions(user, commID, path)
	if allow == false || err != nil {
		if err != nil {
			return nil, "", err
		}
		return nil, "", errors.New("Permission was denied to this file")
	}

	return &Session{
		CommunityID: commID,
		User:        user,
	}, filepath, nil

}

// CommunityFromID gets the whole community struct from just the ID
func CommunityFromID(state *entities.State, comm entities.CommunityID) *entities.Community {
	for _, elem := range state.Communities {
		if elem.ID == comm {
			return &elem
		}
	}
	return nil
}

func (fs *FilesystemService) backnetFromCommID(sessID entities.CommunityID) (Backnet, error) {
	var net Backnet
	for id, backnet := range fs.nets {
		if id == sessID {
			net = backnet
		}
	}
	if net == nil {
		return net, errors.New("this community id has no backnet")
	}
	return net, nil
}

// ParseListFiles prints out the list of files and returns
func (fs *FilesystemService) ParseListFiles(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)
	session, filepath, err := fs.DoChecks(args)

	if err != nil {
		fmt.Println(err)
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(session.CommunityID)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = net.ListFiles(filepath)

	if err != nil {
		fmt.Println(err)
		return
	}

}
