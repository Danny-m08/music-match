package server

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/types"
)

const unableToProcessRequestFormat = "Unable to process request from %s: %s"

var setCookies = []string{"username", "token"}

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
	userReq := types.CreateUserRequest{}

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

	hash, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
	if err != nil {
		logging.Error("Unable to generate bcyrpt hash from password: " + err.Error())
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	user := &types.User{
		Name:     userReq.Name,
		Username: userReq.Username,
		Password: string(hash),
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

	server.setUserCookies(w, user)
	w.WriteHeader(http.StatusOK)
}

func (server *server) uploadFile(w http.ResponseWriter, req *http.Request) {

	username := server.verifyUser(req)
	if username != "" {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		logging.Error(fmt.Sprintf("Error retrieving file from %s: %s", req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	defer file.Close()

	logging.Info("Uploading " + header.Filename)
	dst, err := os.Create(header.Filename)
	if err != nil {
		logging.Error(fmt.Sprintf("Error retrieving file %s from %s: %s", header.Filename, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		logging.Error(fmt.Sprintf("Error retrieving file %s from %s: %s", header.Filename, req.RemoteAddr, err.Error()))
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	logging.Info("File upload successfully")
	w.WriteHeader(http.StatusOK)
}

func (server *server) getUserInfo(w http.ResponseWriter, req *http.Request) {
	userReq := types.GetUserRequest{}

	username := server.verifyUser(req)
	if username != "" {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Error("Unable to process getUserRequest request")
		http.Error(w, "Unable to proccess request", http.StatusBadRequest)
		return
	}

	logging.Info(fmt.Sprintf("get user info request with body %s received", string(body)))

	err = json.Unmarshal(body, &userReq)
	if err != nil {
		logging.Error("Unable to process getUserRequest request: " + err.Error())
		http.Error(w, "Unable to proccess request", http.StatusBadRequest)
		return
	}

	usr := types.User{
		Email:    userReq.Login,
		Username: userReq.Login,
	}

	logging.Info(fmt.Sprintf("Fetching user %s from DB", usr.String()))

	user, err := server.neo4jClient.GetUser(&usr)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to get user %s from DB: %s", user.String(), err.Error()))
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	user.Password = ""
	logging.Info(fmt.Sprintf("Returning %s to client", user.String()))

	resp, err := json.Marshal(user)
	if err != nil {
		logging.Error(fmt.Sprintf("Unable to marshal user %s: %s", user.String(), err.Error()))
		http.Error(w, "Unable to retrieve user info at this time. Please try again later", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		logging.Error(fmt.Sprintf("Error writing response back to %s: %s", req.RemoteAddr, err.Error()))
	}
}

func (server *server) login(w http.ResponseWriter, req *http.Request) {
	loginReq := types.LoginRequest{}

	logging.Info("New login request from " + req.RemoteAddr)

	username := server.verifyUser(req)
	if username != "" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		logging.Error("Error retrieving user from DB: " + err.Error())
		http.Error(w, "Unable to retrieve user's at this time. Please try again later", http.StatusInternalServerError)
		return
	}

	if userInfo == nil {
		logging.Info("No such user found for " + loginReq.Login)
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(loginReq.Password))
	if err != nil {
		logging.Info(fmt.Sprintf("Invalid credentials for user %s: %s ", loginReq.Login, err.Error()))
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		return
	}

	userInfo.Password = ""

	resp, err := json.Marshal(userInfo)
	if err != nil {
		logging.Error("Unable to send userInfo in response: " + err.Error())
		http.Error(w, "Unable to process request. Please try again later", http.StatusInternalServerError)
		return
	}

	logging.Info(fmt.Sprintf("Sending %s back to user", userInfo.String()))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		logging.Error(fmt.Sprintf("Error writing response back to %s: %s", req.RemoteAddr, err.Error()))
	}
}

func (server *server) follow(w http.ResponseWriter, req *http.Request) {
	request := &types.FollowRequest{}

	username := server.verifyUser(req)
	if username != "" {
		http.Error(w, "Session timed out", http.StatusUnauthorized)
		return
	}

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

func (server *server) LogOut(w http.ResponseWriter, req *http.Request) {
	username := server.verifyUser(req)
	if username != "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	delete(server.sessions, username)
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

func (s *server) verifyUser(req *http.Request) string {
	username, err := req.Cookie("username")
	if err != nil || username.Value == "" {
		return ""
	}

	token, ok := s.sessions[username.Value]
	if !ok || token == "" {
		return ""
	}

	cookieToken, err := req.Cookie("token")
	if err != nil || cookieToken.Value != token {
		return ""
	}

	return username.Value
}

func (s *server) setUserCookies(w http.ResponseWriter, user *types.User) {
	if user.Username == "" {
		return
	}

	const alphaNumeric = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJLKMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	token := ""
	for it := 0; it < 32; it++ {
		token += string(alphaNumeric[rand.Intn(len(alphaNumeric))])
	}

	s.sessions[user.Username] = token

	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: user.Username,
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
	})

}
