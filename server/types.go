package server

import (
	"github.com/danny-m08/music-match/types"
)

type followRequest struct {
	User     *types.User `json:"user" :"user"`
	Follower *types.User `json:"follower" :"follower"`
}

type followers struct {
	followers []*types.User `json:"followers"`
}
