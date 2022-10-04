package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/types"
)

const unableToProcessRequestFormat = "Unable to process request from %s: %s"

func (s *server) newUser(w http.ResponseWriter, req *http.Request) {
	logging.Info("New user request from: " + req.RemoteAddr)

	body, err := io.ReadAll(req.Body)
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
		Username: userReq.Username,
		Password: userReq.Password,
		Email:    userReq.Email,
	}

	logging.Info("Creating new user " + user.String())
	err = s.neo4jClient.InsertUser(user)
	if err != nil {
		logging.Error(fmt.Sprintf(unableToProcessRequestFormat, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to create new user", http.StatusInternalServerError)
		return
	}

	token, err := s.createJWT(user)
	if err != nil {
		logging.Error("Error creating token: " + err.Error())
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}

	body, err = s.getTokenPayload(token)
	if err != nil {
		logging.Error("Error creating payload: " + err.Error())
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}

	logging.Info(fmt.Sprintf("%s created successfully", user.String()))
	w.Write(body)
}

func (s *server) login(w http.ResponseWriter, req *http.Request) {
	loginReq := LoginRequest{}

	logging.Debug("Checking credentials for %s before continuing login process")
	valid, err := s.checkCredentials(req)
	if err != nil {
		logging.Info("Unable to validate credentials: " + err.Error())
	} else if !valid {
		logging.Info("Invalid credentials, continuing login process")
	} else if valid {
		logging.Info("Valid token!")
		w.WriteHeader(http.StatusOK)
	}

	body, err := io.ReadAll(req.Body)
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

	userInfo, err := s.neo4jClient.GetUser(&usr)
	if err != nil {
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
	}

	if userInfo.Password == loginReq.Password {
		http.Error(w, "Username or password incorrect", http.StatusUnauthorized)
	}

	token, err := s.createJWT(&usr)
	if err != nil {
		logging.Error("Error creating token: " + err.Error())
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}

	body, err = s.getTokenPayload(token)
	if err != nil {
		logging.Error("Error creating payload: " + err.Error())
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}

	logging.Info("Successful signing from " + usr.String())
	w.Write(body)
}

func (s *server) follow(w http.ResponseWriter, req *http.Request) {
	request := &followRequest{}

	logging.Debug("Checking credentials for %s")
	valid, err := s.checkCredentials(req)
	if err != nil {
		logging.Info("Unable to validate credentials: " + err.Error())
		w.WriteHeader(http.StatusUnauthorized)
	} else if !valid {
		logging.Info("Invalid token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Unable to process request: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, request)
	if err != nil {
		http.Error(w, "Unable to process request: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = s.neo4jClient.CreateFollowing(request.User, request.Follower)
	if err != nil {
		logging.Error("Unable to create following request: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	logging.Info(fmt.Sprintf("%s -> %s following created successfully", request.Follower.String(), request.User.String()))
}

func (s *server) getFollowers(w http.ResponseWriter, req *http.Request) {
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
