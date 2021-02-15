package client

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/eagraf/habitat-node/entities"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles requests and performs business logic
type UserService struct {
	userRepo *UserRepo
}

// UserRepo handles interaction with the database layer for Users
// Currently the db layer is just a json file. When that gets refactored, only changes
// will need to be made to UserRepo, the rest of the User logic will not be touched.
type UserRepo struct {
	path  string
	mutex sync.Mutex
}

// HostUser represents a user registered on this client
type HostUser struct {
	ID          int         `json:"id"`
	Name        string      `json:"username"`
	Hash        string      `json:"hash"`
	Permissions Permissions `json:"permissions"`
	Spaces      map[int]entities.Community
}

// User is HostUser renamed for clarity/convenience
type User HostUser

// UserTable is how users are stored in the file db
// why isn't this just a map of username to user ?
type UserTable struct {
	Users         map[int]*User  `json:"users"`
	UsernameIndex map[string]int `json:"usersname_index`
}

// Permissions belonging to a user
type Permissions struct {
	Admin bool `json:"admin"`
}

// CreateUserRequest is the body expected by CreateUserHandler
type CreateUserRequest struct {
	Name        string      `json:"username"`
	Password    string      `json:"password"`
	Permissions Permissions `json:"Permissions"`
}

// NewUserService creates a new UserService
func NewUserService(userRepo *UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUserHandler handles POST to create a new user on this node
// Only works if the user is an admin
func (us *UserService) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Check to see if user is an admin
	user, err := getUserFromContextHelper(r.Context())
	if err != nil {
		http.Error(w, "user decode failed", http.StatusInternalServerError)
		return
	}

	if user.Permissions.Admin == false {
		http.Error(w, "user is not an admin", http.StatusForbidden)
		return
	}

	// Read request body and validate
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	var body CreateUserRequest
	err = json.Unmarshal(buf, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = us.addUser(body.Name, body.Password, body.Permissions)
	if err != nil {
		http.Error(w, "failed to add user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// BootstrapUserHandler handles the creation of the first user
func (us *UserService) BootstrapUserHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body and validate
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	var body CreateUserRequest
	err = json.Unmarshal(buf, &body)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	setup, err := us.isSetup()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !setup {
		http.Error(w, "not in bootstrap mode", http.StatusBadRequest)
		return
	}

	// Create new admin user
	us.addUser(body.Name, body.Password, Permissions{
		Admin: true,
	})

	w.WriteHeader(http.StatusOK)
}

func (us *UserService) addUser(username, password string, permissions Permissions) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &User{
		Name:        username,
		Hash:        string(hash),
		Permissions: permissions,
	}
	err = us.userRepo.CreateUser(user)
	if err != nil {
		return err
	}
	return nil
}

// If there are no users in the system, the client is in setup mode
func (us *UserService) isSetup() (bool, error) {
	users, err := us.userRepo.GetAllUsers()
	if err != nil {
		return false, err
	}
	return len(users) == 0, nil
}

// NewUserRepo creates a new UserRepo
func NewUserRepo(repoDir string) (*UserRepo, error) {
	ur := &UserRepo{
		path:  filepath.Join(repoDir, "users"),
		mutex: sync.Mutex{},
	}

	err := ur.writeUserFile(&UserTable{
		Users:         make(map[int]*User),
		UsernameIndex: make(map[string]int),
	})
	if err != nil {
		return nil, err
	}

	return ur, nil
}

// CreateUser adds a user to the db
func (ur *UserRepo) CreateUser(user *User) error {
	users, err := ur.readUserFile()
	if err != nil {
		return err
	}

	if _, taken := users.UsernameIndex[user.Name]; taken {
		return errors.New("that username is already taken")
	}

	user.ID = len(users.Users)
	users.Users[user.ID] = user
	users.UsernameIndex[user.Name] = user.ID

	err = ur.writeUserFile(users)
	if err != nil {
		return err
	}
	return nil
}

// GetUser retrieves a user from the db
func (ur *UserRepo) GetUser(id int) (*User, error) {
	users, err := ur.readUserFile()
	if err != nil {
		return nil, err
	}

	if user, ok := users.Users[id]; ok {
		return user, nil
	}
	return nil, errors.New("user does not exist")
}

// GetUserByUsername gets a user by their screen name
func (ur *UserRepo) GetUserByUsername(username string) (*User, error) {
	users, err := ur.readUserFile()
	if err != nil {
		return nil, err
	}

	if userID, ok := users.UsernameIndex[username]; ok {
		if user, ok := users.Users[userID]; ok {
			return user, nil
		}
		return nil, errors.New("failed to get user")
	}
	return nil, errors.New("user does not exist")
}

// GetAllUsers returns all users in the db
func (ur *UserRepo) GetAllUsers() (map[int]*User, error) {
	users, err := ur.readUserFile()
	if err != nil {
		return nil, err
	}
	return users.Users, nil
}

func (ur *UserRepo) readUserFile() (*UserTable, error) {
	ur.mutex.Lock()
	defer ur.mutex.Unlock()

	buf, err := ioutil.ReadFile(ur.path)
	if err != nil {
		return nil, err
	}

	var users UserTable
	err = json.Unmarshal(buf, &users)
	if err != nil {
		return nil, err
	}

	return &users, nil
}

func (ur *UserRepo) writeUserFile(users *UserTable) error {
	ur.mutex.Lock()
	defer ur.mutex.Unlock()

	buf, err := json.Marshal(users)
	if err != nil {
		return err
	}

	// log.Print("here ", ur.path)
	err = ioutil.WriteFile(ur.path, buf, 0600)
	if err != nil {
		return err
	}
	return nil
}
