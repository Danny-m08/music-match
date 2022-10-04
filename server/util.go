package server

import (
	"encoding/json"
	"errors"
	"github.com/danny-m08/music-match/types"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

type UserClaims struct {
	*jwt.RegisteredClaims
	User UserAuth `json:"login"`
}

type UserAuth struct {
	login string
}

type tokenPayload struct {
	token string `json:"token""`
}

var expiry = time.Hour

// createJWT creates a token and returns it
func (s *server) createJWT(user *types.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	usr := UserAuth{}

	if user.Email != "" {
		usr.login = user.Email
	} else if user.Username != "" {
		usr.login = user.Password
	} else {
		return "", errors.New("username or password must be provided in order to generate a JWT")
	}

	token.Claims = UserClaims{
		&jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		usr,
	}

	return token.SignedString(s.secret)
}

// validateJWT validates a token and returns whether it is valid or not
func (s *server) validateJWT(token string) (bool, error) {
	parser := jwt.NewParser()

	t, err := parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return false, err
	}

	return t.Valid, nil
}

func (s *server) checkCredentials(r *http.Request) (bool, error) {
	return s.validateJWT(r.Header.Get("Authorization"))
}

func (s *server) getTokenPayload(token string) ([]byte, error) {
	return json.Marshal(tokenPayload{token: token})
}
