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
	logging.Info("Responding to client with 200 OK")
	_, err = w.Write([]byte("Welcome!"))
	if err != nil {
		logging.Error("Unable to write back to the client: " + err.Error())
	}

}

func (server *server) getUserInfo(w http.ResponseWriter, req *http.Request) {
	userReq := GetUserRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Error("Unable to process getUserRequest request")
		http.Error(w, "Unable to proccess request", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &userReq)
	if err != nil {
		logging.Error("Unable to process getUserRequest request: " + err.Error())
		http.Error(w, "Unable to proccess request", http.StatusBadRequest)
		return
	}

	usr := types.User{
		Email:    userReq.Login,
		Password: userReq.Login,
	}

	user, err := server.neo4jClient.GetUser(&usr)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to get user %s from DB: %s", user.String(), err.Error()))
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	user.Password = ""

	resp, err := json.Marshal(user)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to marshal user %s: %s", user.String(), err.Error()))
		http.Error(w, "Unable to retrieve user info at this time. Please try again later", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
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
		Email:    loginReq.Login,
		Username: loginReq.Login,
		Password: loginReq.Password,
	}

	logging.Info("Checking for user " + usr.String())

	userInfo, err := server.neo4jClient.GetUser(&usr)
	if err != nil {
		http.Error(w, "Unable to retrieve user's at this time. Please try again later", http.StatusUnauthorized)
		return
	}

	if userInfo == nil {
		logging.Info("No such user found for " + loginReq.Login)
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		return
	} else if userInfo.Password != loginReq.Password {
		logging.Info("Invalid credentials for user " + loginReq.Login)
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		return
	}

	logging.Info("User data successfully retrieved from DB for" + loginReq.Login)
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
