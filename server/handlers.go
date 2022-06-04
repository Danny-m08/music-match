package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/neo4j"
)

var handlers = map[string]func(http.ResponseWriter, *http.Request){
	"/newuser":   newUser,
	"/follow":    follow,
	"/followers": getFollowers,
	"/login":     login,
}

func newUser(w http.ResponseWriter, req *http.Request) {
	user, err := readUser(req)
	if err != nil {
		logging.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logging.Info("New user request from: " + req.RemoteAddr)

	client, err := neo4j.NewClient(config.GetGlobalConfig().GetDBConfig())
	if err != nil {
		logging.Error(err.Error())
		http.Error(w, "Internal server error please try again later", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	logging.Info("Creating new user " + user.String())

	err = client.InsertUser(user)
	if err != nil {
		logging.Error(err.Error())
		http.Error(w, "Unable to create new user:"+err.Error(), http.StatusInternalServerError)
		return
	}

	logging.Info(fmt.Sprintf("%s created successfully", user.String()))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success!"))
}

func login(w http.ResponseWriter, req *http.Request) {
	user, err := readUser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	client, err := neo4j.NewClient(config.GetGlobalConfig().GetDBConfig())
	if err != nil {
		http.Error(w, "Internal server error please try again later", http.StatusInternalServerError)
	}
	defer client.Close()

	userInfo, err := client.GetUser(user)
	if err != nil {
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
	}

	if userInfo.Password == user.Password {
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
	}

	w.Write([]byte("Successful login!"))
	w.WriteHeader(http.StatusOK)
}

func follow(w http.ResponseWriter, req *http.Request) {
	request := &followRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Unable to process request: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, request)
	if err != nil {
		http.Error(w, "Unable to process request: "+err.Error(), http.StatusBadRequest)
		return
	}

	client, err := neo4j.NewClient(config.GetGlobalConfig().GetDBConfig())
	if err != nil {
		logging.Error("Unable to create DB connection: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	err = client.CreateFollowing(request.User, request.Follower)
	if err != nil {
		logging.Error("Unable to create following request: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	logging.Info(fmt.Sprintf("%s -> %s following created successfully", request.Follower.String(), request.User.String()))
}

func getFollowers(w http.ResponseWriter, req *http.Request) {
	user, err := readUser(req)
	if err != nil {
		logging.Error("Unable to read user from request: " + err.Error())
		http.Error(w, "Unable to proccess request: "+err.Error(), http.StatusBadRequest)
		return
	}

	client, err := neo4j.NewClient(config.GetGlobalConfig().GetDBConfig())
	if err != nil {
		logging.Error("Unable to create DB connection: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	f, err := client.GetFollowers(user)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to retrieve followers for user %s: %s", user.String(), err.Error()))
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
	}

	fReq := followers{
		followers: f,
	}

	data, err := json.Marshal(fReq)
	if err != nil {
		logging.Error("Unable to marshal followers: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
	}

	w.Write(data)
}
