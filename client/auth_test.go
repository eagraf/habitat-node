package client

import (
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestGrantVerifyJWT(t *testing.T) {
	user := &User{
		ID:   0,
		Name: "Alice",
		Permissions: Permissions{
			Admin: true,
		},
	}

	token, err := grantToken(user)
	assert.Nil(t, err)

	parsed, err := verifyToken(token)
	assert.Nil(t, err)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, claims["authorized"].(bool))
	assert.Equal(t, 0, int(claims["user_id"].(float64)))
	assert.Equal(t, true, claims["permissions"].(map[string]interface{})["admin"])
}
