package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles all authentication related tasks for the client
type AuthService struct {
	tokenRepo *TokenRepo
	userRepo  *UserRepo
}

// TokenRepo handles interaction with the database layer for tokens
type TokenRepo struct {
	path  string
	mutex sync.Mutex
}

// LoginRequest is the body expected by LoginHandler
type LoginRequest struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the body returned by LoginHandler
type LoginResponse struct {
	Token string `json:"auth_token"`
}

type ctxKey int

const (
	ctxKeyUser ctxKey = 0
)

// NewAuthService initializes a new auth service
func NewAuthService(tr *TokenRepo, ur *UserRepo) (*AuthService, error) {
	authPath := os.Getenv("AUTH_DIR")
	if authPath == "" {
		return nil, errors.New("AUTH_DIR env var must be set")
	}

	res := &AuthService{
		tokenRepo: tr,
		userRepo:  ur,
	}

	return res, nil
}

// Middleware fulfills the gorilla/mux Middleware interface
func (as *AuthService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		user, err := as.CheckToken(token)
		if err != nil {
			http.Error(w, "unathorized", http.StatusUnauthorized)
		} else {
			log.Printf("Authenticated user %s", user.Name)
			ctx := context.WithValue(r.Context(), ctxKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// LoginHandler lets a user login
func (as *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var loginReq LoginRequest
	err = json.Unmarshal(buf, &loginReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := as.checkUser(loginReq.Name, loginReq.Password)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Grant a token if username and password match
	token, err := grantToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Save the token in the tokens file
	err = as.tokenRepo.CreateToken(user.ID, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := LoginResponse{Token: token}
	buf, err = json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(buf)
}

// LogoutHandler expires a users token
func (as *AuthService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContextHelper(r.Context())
	if err != nil {
		http.Error(w, "user decode failed", http.StatusInternalServerError)
		return
	}
	err = as.tokenRepo.DeleteToken(user.ID)
	if err != nil {
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (as *AuthService) checkUser(username, password string) (*User, error) {

	user, err := as.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password))
	if err != nil {
		return nil, errors.New("incorrect password")
	}
	return user, nil
}

// CheckToken checks a token and returns the associated user
func (as *AuthService) CheckToken(token string) (*User, error) {
	parsed, err := verifyToken(token)
	if err != nil {
		return nil, err
	}

	// Match token to stored tokens
	claimsMap := parsed.Claims.(jwt.MapClaims)
	userID := int(claimsMap["user_id"].(float64))

	storedToken, err := as.tokenRepo.GetToken(userID)
	if err != nil {
		return nil, err
	}
	if storedToken != token {
		return nil, errors.New("unrecognized auth token")
	}

	var permissions Permissions
	permissionsMap := claimsMap["permissions"].(map[string]interface{})
	mapstructure.Decode(permissionsMap, &permissions)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:          userID,
		Name:        claimsMap["username"].(string),
		Permissions: permissions,
	}, nil
}

// NewTokenRepo creates a new TokenRepo
func NewTokenRepo(repoDir string) (*TokenRepo, error) {
	tr := &TokenRepo{
		path:  filepath.Join(repoDir, "tokens"),
		mutex: sync.Mutex{},
	}

	err := tr.writeTokenFile(make(map[int]string))
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// CreateToken creates or replaces an existing token
func (tr *TokenRepo) CreateToken(userID int, token string) error {
	tokens, err := tr.readTokenFile()
	if err != nil {
		return err
	}

	tokens[userID] = token
	err = tr.writeTokenFile(tokens)
	if err != nil {
		return err
	}
	return nil
}

// GetToken gets a token by user id
func (tr *TokenRepo) GetToken(userID int) (string, error) {
	tokens, err := tr.readTokenFile()
	if err != nil {
		return "", err
	}

	if token, exists := tokens[userID]; exists {
		return token, nil
	}
	return "", errors.New("token not found")
}

// DeleteToken clears a token from the db
func (tr *TokenRepo) DeleteToken(userID int) error {
	tokens, err := tr.readTokenFile()
	if err != nil {
		return err
	}

	if _, ok := tokens[userID]; ok {
		delete(tokens, userID)
	}
	err = tr.writeTokenFile(tokens)
	if err != nil {
		return err
	}
	return nil
}

func (tr *TokenRepo) readTokenFile() (map[int]string, error) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	buf, err := ioutil.ReadFile(tr.path)
	if err != nil {
		return nil, err
	}

	var tokens map[int]string
	err = json.Unmarshal(buf, &tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (tr *TokenRepo) writeTokenFile(tokens map[int]string) error {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	buf, err := json.Marshal(tokens)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(tr.path, buf, 0600)
	if err != nil {
		return err
	}
	return nil
}

func grantToken(user *User) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = user.ID
	claims["username"] = user.Name
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["permissions"] = user.Permissions

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return signed, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	// Check that the token is still valid
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}
	return token, err
}

func getUserFromContextHelper(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(ctxKeyUser).(*User)
	if !ok {
		return nil, errors.New("failed to read user from context")
	}
	return user, nil
}
