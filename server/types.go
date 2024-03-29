package server

import (
	"github.com/danny-m08/music-match/types"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type followRequest struct {
	User     *types.User `json:"user"`
	Follower *types.User `json:"follower"`
}
