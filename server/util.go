package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/danny-m08/music-match/types"
)

//readUser creates a user object from the http request body
func readUser(req *http.Request) (*types.User, error) {
	user := &types.User{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
