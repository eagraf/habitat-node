package fs

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
)

// Session is a logged in user
// TODO: move to somewhere else?
type Session struct {
	CommunityID entities.CommunityID
	User        *client.User // this is the "wrong" type of user
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
	fmt.Print("path to parse ", path, "\n")

	tomatch := `[a-zA-Z0-9_-]*:/*[^/]*.*`

	if ok, _ := regexp.MatchString(tomatch, path); !ok {
		return "", "", errors.New("the given path did not pass regex")
	}

	arr := strings.Split(path, ":")
	fmt.Println("comm: ", arr[0], " path ", arr[1])

	if len(arr) < 2 {
		return "", "", errors.New("invalid path format")
	}

	if arr[0] == "" {
		return "", "", errors.New("no comm id specified")
	}

	if arr[1] == "" {
		arr[1] = "."
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
	if err != nil {
		return nil, "", err
	}

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

// DoChecksSimple is for use before we have user authentication set up
// here we pass in a community id (rather than session token) and assume user has permissions
// since we don't have users set up yet
func (fs *FilesystemService) DoChecksSimple(args url.Values) (entities.CommunityID, string, error) {

	path := args.Get("path")
	if path == "" {
		return "", "", errors.New("Empty path not accepted")
	}

	commID, filepath, err := ParseFilePath(path)
	if err != nil {
		return "", "", err
	}

	return commID, filepath, nil

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

// GetBacknetFromRequest is a helper function for things pre-IPFS HTTP API calls
/*
func (fs *FilesystemService) GetBacknetFromRequest(args url.Values) (Backnet, error) {

}
*/

// ParseListFiles prints out the list of files and returns
func (fs *FilesystemService) ParseListFiles(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)
	// session, filepath, err := fs.DoChecks(args) // for when users are set up

	commid, filepath, err := fs.DoChecksSimple(args)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(commid)
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

// ParseWrites handles requests to write files
func (fs *FilesystemService) ParseWrites(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	commid, filepath, err := fs.DoChecksSimple(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	fname := args.Get("file")
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(commid)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = net.Write(filepath, file)

	if err != nil {
		fmt.Println(err)
		return
	}

}

// ParsePinActions handles checks for is pins, pinning and unpinning
func (fs *FilesystemService) ParsePinActions(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	commid, filepath, err := fs.DoChecksSimple(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	action := args.Get("action")

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(commid)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch action {
	case "check":
		_, err = net.IsPinned(filepath)
	case "pin":
		err = net.Pin(filepath)
	case "unpin":
		err = net.Unpin(filepath)
	}

	if err != nil {
		fmt.Println(err)
		return
	}

}

// ParseRemoves handles requests to remove files
func (fs *FilesystemService) ParseRemoves(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	commid, filepath, err := fs.DoChecksSimple(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	isdirstr := args.Get("isdir")
	isdir := false
	if isdirstr == "true" {
		isdir = true
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(commid)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = net.Remove(filepath, bool(isdir))

	if err != nil {
		fmt.Println(err)
		return
	}

}

// ParseCats handles requests to cat files
func (fs *FilesystemService) ParseCats(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	commid, filepath, err := fs.DoChecksSimple(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(commid)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := net.Cat(filepath)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(res))
	}

}

// ParseMoves handles requests to move files
func (fs *FilesystemService) ParseMoves(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	oldpath := args.Get("old")
	if oldpath == "" {
		fmt.Println("oldpath cannot be an empty argument")
		return
	}

	newpath := args.Get("new")
	if newpath == "" {
		fmt.Println("newpath cannot be an empty argument")
		return
	}

	oldcommID, oldfilepath, err := ParseFilePath(oldpath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	newcommID, newfilepath, err := ParseFilePath(newpath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if oldcommID != newcommID {
		fmt.Println("cannot move files between communities (yet!)")
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(oldcommID)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = net.Move(oldfilepath, newfilepath)

	if err != nil {
		fmt.Println(err)
	}
}

// ParseCopys handles requests to copy files
func (fs *FilesystemService) ParseCopys(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	oldpath := args.Get("old")
	if oldpath == "" {
		fmt.Println("oldpath cannot be an empty argument")
	}

	newpath := args.Get("new")
	if newpath == "" {
		fmt.Println("newpath cannot be an empty argument")
		return
	}

	oldcommID, oldfilepath, err := ParseFilePath(oldpath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	newcommID, newfilepath, err := ParseFilePath(newpath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if oldcommID != newcommID {
		fmt.Println("cannot move files between communities (yet!)")
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(oldcommID)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = net.Copy(oldfilepath, newfilepath)

	if err != nil {
		fmt.Println(err)
	}
}

// ParseMkdirs handles requests to make a new directory
func (fs *FilesystemService) ParseMkdirs(w http.ResponseWriter, r *http.Request) {

	args := GetRequestQueries(r)

	commid, filepath, err := fs.DoChecksSimple(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	// how to get community backnet from user
	net, err := fs.backnetFromCommID(commid)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = net.MakeDir(filepath)

	if err != nil {
		fmt.Println(err)
	}

}
