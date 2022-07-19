package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/types"
)

const unableToProcessRequestFormat = "Unable to process request from %s: %s"

func (server *server) newUser(w http.ResponseWriter, req *http.Request) {
	logging.Info("New user request from: " + req.RemoteAddr)
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Error(fmt.Sprintf(unableToProcessRequestFormat, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	logging.Debug(string(body))
	userReq := CreateUserRequest{}

	err = json.Unmarshal(body, &userReq)
	if err != nil {
		logging.Error(fmt.Sprintf(unableToProcessRequestFormat, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	valid, err := verifyEmail(userReq.Email)
	if err != nil {
		logging.Error(fmt.Sprintf(unableToProcessRequestFormat, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	} else if !valid {
		logging.Error(fmt.Sprintf(unableToProcessRequestFormat, req.RemoteAddr, "Invalid email address "+userReq.Email))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	user := &types.User{
		Name:     userReq.Name,
		Username: userReq.Username,
		Password: userReq.Password,
		Email:    userReq.Email,
	}

	logging.Info("Creating new user " + user.String())
	err = server.neo4jClient.InsertUser(user)
	if err != nil {
		logging.Error(fmt.Sprintf(unableToProcessRequestFormat, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to create new user", http.StatusInternalServerError)
		return
	}

	logging.Info(fmt.Sprintf("%s created successfully", user.String()))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Welcome!"))
	if err != nil {
		logging.Error("Unable to write back to the client: " + err.Error())
	}

}

func (server *server) login(w http.ResponseWriter, req *http.Request) {
	loginReq := LoginRequest{}

	logging.Info("New login request from " + req.RemoteAddr)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &loginReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usr := types.User{
		Email:    loginReq.Email,
		Password: loginReq.Password,
	}

	logging.Info("Checking for user " + usr.Email)

	userInfo, err := server.neo4jClient.GetUser(&usr)
	if err != nil {
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
		return
	}

	if userInfo == nil || userInfo.Password != loginReq.Password {
		logging.Info("Invalid credentials for user " + usr.Email)
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (server *server) follow(w http.ResponseWriter, req *http.Request) {
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

	err = server.neo4jClient.CreateFollowing(request.User, request.Follower)
	if err != nil {
		logging.Error("Unable to create following request: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	logging.Info(fmt.Sprintf("%s -> %s following created successfully", request.Follower.String(), request.User.String()))
}

func (server *server) getFollowers(w http.ResponseWriter, req *http.Request) {
	//user, err := readUser(req)
	//if err != nil {
	//	logging.Error("Unable to read user from request: " + err.Error())
	//	http.Error(w, "Unable to proccess request: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	//	client, err := neo4j.NewClient(config.GetGlobalConfig().GetDBConfig())
	//	if err != nil {
	//		logging.Error("Unable to create DB connection: " + err.Error())
	//		http.Error(w, "Unable to process request", http.StatusInternalServerError)
	//		return
	//	}
	//	defer client.Close()
	//
	//	f, err := client.GetFollowers(user)
	//	if err != nil {
	//		logging.Error(fmt.Sprintf("Unable to retrieve followers for user %s: %s", user.String(), err.Error()))
	//		http.Error(w, "Unable to process request", http.StatusInternalServerError)
	//	}
	//
	//	fReq := followers{
	//		followers: f,
	//	}
	//
	//	data, err := json.Marshal(fReq)
	//	if err != nil {
	//		logging.Error("Unable to marshal followers: " + err.Error())
	//		http.Error(w, "Unable to process request", http.StatusInternalServerError)
	//	}
	//
	//	w.Write(data)
}
